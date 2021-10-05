package main

import (
	"context"
	"example.com/itsuMain/lib/connection"
	"log"
)

func main() {
	var err error
	var listener connection.Listener

	server := NewServer()

	if listener, err = connection.NewListener("0.0.0.0:15184"); err != nil {
		log.Panicln(err)
	}

	for {
		var s connection.Session
		if s, err = connection.Accept(listener, context.Background()); err != nil {
			log.Println(err)
			continue
		}

		var c *Client
		if c, err = server.NewClient(s); err != nil {
			log.Println("Error while accepting a new client:", err)
		}
		log.Println("Accepted new client with ID", c.identifier)
	}
}
