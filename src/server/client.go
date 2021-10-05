package main

import (
	"example.com/itsuMain/lib/connection"
	"example.com/itsuMain/lib/util"
	"fmt"
	"log"
	"sync"
)

type Client struct {
	Session connection.Session

	identifier uint64
	sysInfo    util.SystemInformation

	currentToken uint64

	threadsWG *sync.WaitGroup
	done      bool //for gc, async reads that have race conditions doesn't matter
}

type clientLogger struct {
	identifier uint64
}

func (c clientLogger) println(v ...interface{}) {
	log.Printf("[%d]: %s", c.identifier, fmt.Sprint(v...))
}

func (c *Client) logger() clientLogger {
	return clientLogger{identifier: c.identifier}
}

func (c *Client) Main() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Client worker for id", c.identifier, "recovered:", r)
		}
	}()

	defer func() {
		c.threadsWG.Done()
		c.threadsWG.Wait()
		c.done = true
	}()

	logger := c.logger()

	logger.println("worker started")

	for {
		if msg, _, err := c.Session.ReadMessage(); err != nil {
			logger.println("couldn't read message: ", err)
			break
		} else {
			log.Println(msg)
		}
	}

	logger.println("worker stopping")
}
