package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

var h Hub

type Hub struct {
	connections map[*Connection]bool
	register    chan *Connection
	unregister  chan *Connection
	refresh     chan bool
	onRefresh   func() []byte
}

func NewHub(f func() []byte) {
	h = Hub{
		refresh:     make(chan bool),
		onRefresh:   f,
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		connections: make(map[*Connection]bool),
	}
}

func Register(c *Connection) {
	h.register <- c
}

func Unregister(c *Connection) {
	h.unregister <- c
}

func Refresh() {
	h.refresh <- true
}

func HubLoop() {
	for {
		select {
		case c := <-h.register:
			fmt.Println("Connect")
			h.connections[c] = true
			bytes := h.onRefresh()
			c.send <- bytes
		case c := <-h.unregister:
			fmt.Println("Disconnect")
			delete(h.connections, c)
			close(c.send)
		case <-h.refresh:
			fmt.Println("Refreshing")
			bytes := h.onRefresh()
			for c := range h.connections {
				select {
				case c.send <- bytes:
				default:
					delete(h.connections, c)
					close(c.send)
					go c.ws.Close()
				}
			}
		}
	}
}

// Goroutine wrapper for a websocket connection. Anything sent on the `send` channel
// will be written to the websocket.
type Connection struct {
	ws   *websocket.Conn
	send chan []byte
}

// Creates and returns a new Connection object
func NewConnection(ws *websocket.Conn) *Connection {
	return &Connection{send: make(chan []byte, 256), ws: ws}
}

// Listens forever on the websocket, performing actions as needed.
func (c *Connection) Reader() {
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			fmt.Printf("Error in websocket read: %s\n", err.Error())
			break
		}
		fmt.Printf("Message received: %s\n", msg)
		// TODO: Do something with the msg
	}
	c.ws.Close()
}

// Writes anything on the send channel to the websocket.
func (c *Connection) Writer() {
	for msg := range c.send {
		err := c.ws.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			fmt.Printf("Error in websocket write: %s\n", err.Error())
			break
		}
	}
	c.ws.Close()
}