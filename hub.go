// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "sync"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.

type Channel struct {
	in    chan []byte
	out   chan []byte
	state string
}

type Hub struct {
	// Registered clients.
	clients  map[*Client]bool
	supports map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	chanMu   sync.Mutex
	channels []Channel

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast: make(chan []byte),
		channels:  make([]Channel, 0),

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
					delete(h.supports, client)
					close(client.send)
				}
			default:
				if _, ok := h.clients[client]; ok {
					delete(h.clients, client)
					close(client.send)
				}
			}

			// case message := <-h.broadcast:
			// 	for client := range h.clients {
			// 		select {
			// 		case client.send <- message:
			// 		default:
			// 			close(client.send)
			// 			delete(h.clients, client)
			// 		}
			// 	}
		}
	}
}
