package main

import (
	"fmt"
	"sync"
)

type Channel struct {
	in    chan []byte
	out   chan []byte
	state string
}

type Hub struct {
	clients  map[*Client]bool
	supports map[*Client]bool

	broadcast chan []byte

	chanMu   sync.Mutex
	channels []*Channel

	register chan *Client

	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast: make(chan []byte),
		channels:  make([]*Channel, 0),

		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		supports:   make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			if client.role == "support" {
				h.supports[client] = true
			} else {
				h.clients[client] = true
			}
		case client := <-h.unregister:
			switch client.role {
			case "support":
				if _, ok := h.supports[client]; ok {
					fmt.Println("Support logged out")
					client.hub.chanMu.Lock()
					delete(h.supports, client)
					close(client.channel.in)
				}
			default:
				if _, ok := h.clients[client]; ok {
					fmt.Println("Client logged out")
					// client.channel.state = "Close"
					delete(h.clients, client)
					close(client.channel.in)
				}
			}
		}
	}
}
