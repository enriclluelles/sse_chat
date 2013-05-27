package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Message struct {
	Who  string
	What string
}

type ClientList struct {
	clients map[chan *Message]bool
}

func (self *ClientList) AddClient(client chan *Message) {
	//Initialize the clients map if it's not
	if self.clients == nil {
		self.clients = make(map[chan *Message]bool)
	}

	self.clients[client] = true
}

func (self *ClientList) DeleteClient(client chan *Message) {
	if self.clients != nil {
		delete(self.clients, client)
	}
}

func (self *ClientList) BroadCast(message Message) {
	for client, _ := range self.clients {
		client <- &message
	}
}

func main() {

	cl := new(ClientList)

	staticHandler := func(res http.ResponseWriter, req *http.Request) {
		http.ServeFile(res, req, "public"+req.URL.Path)
	}

	postMessageHandler := func(res http.ResponseWriter, req *http.Request) {
		m := Message{
			Who:  req.FormValue("who"),
			What: req.FormValue("what"),
		}
		cl.BroadCast(m)
	}

	sseHandler := func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/event-stream\n\n")

		f, ok := res.(http.Flusher)

		c := make(chan *Message)

		cl.AddClient(c)

		for i := 0; ; i++ {

			m := <-c

			json_response, _ := json.Marshal(m)

			body := fmt.Sprintf("event: said\ndata: %s\n\n", json_response)

			var _, err = res.Write([]byte(body))

			if err != nil {
				break //if we can't write, stop the loop
			}

			if ok {
				f.Flush()
			} else {
				log.Printf("w does not support flush")
			}
		}

		cl.DeleteClient(c)
	}

	http.HandleFunc("/", staticHandler)
	http.HandleFunc("/message", postMessageHandler)
	http.HandleFunc("/events", sseHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("starting the server on port %s", port)

	httpServer := &http.Server{
		Addr: ":" + port,
	}

	httpServer.ListenAndServe()
}
