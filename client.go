// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"2019_2_IBAT/pkg/pkg/interfaces"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	role    string
	channel *Channel
}

func (c *Client) readPump() {
	// defer func() {
	// 	c.hub.unregister <- c
	// 	// fmt.Print("Connection closed")
	// 	// if c.conn.
	// 	// c.conn.Close()
	// }()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		fmt.Printf("User write %s", string(message))
		c.channel.out <- message
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	// defer func() {
	// 	ticker.Stop()
	// 	fmt.Print("Connection closed")
	// 	c.conn.Close()
	// }()

	for {
		select {
		case message, ok := <-c.channel.in:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok && c.role == "support" {
				c.channel.in = make(chan []byte)
				c.channel.state = "Open"
			}

			if !ok && (c.role == "seeker" || c.role == "employer") {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.conn.Close()
				fmt.Println("Channel is closed")
				return
			}

			fmt.Printf("User accepted message %s\n", string(message))
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.channel.in)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.channel.in)
			}

			if err := w.Close(); err != nil {
				fmt.Println("Channel is closed")
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {

	authInfo, ok := interfaces.FromContext(r.Context())
	fmt.Println("User authorized:")
	fmt.Println(authInfo)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	role := authInfo.Role

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, role: role}
	client.hub.register <- client

	if role == "support" {
		client.hub.chanMu.Lock()

		client.channel = &Channel{
			in:    make(chan []byte),
			out:   make(chan []byte),
			state: "Open",
		}

		client.hub.channels = append(client.hub.channels, client.channel)
		client.hub.chanMu.Unlock()
		fmt.Printf("support created")
	}

	if role == "seeker" || role == "employer" {
		client.hub.chanMu.Lock()

		success := false
		for _, ch := range client.hub.channels {
			if ch.state == "Open" {
				ch.state = "Engaged"

				fmt.Printf("Channel state switched, now %s\n", ch.state)
				client.channel = &Channel{
					in:  ch.out,
					out: ch.in,
				}
				success = true
				break
			}
		}
		client.hub.chanMu.Unlock()

		if !success {
			w.WriteHeader(524)
			fmt.Printf("no free supports")
			// fmt.Printf("Clients left %d", len(client.hub.clients))
			return
		} else {
			fmt.Printf("user created")
		}
	}

	go client.writePump()
	go client.readPump()
}
