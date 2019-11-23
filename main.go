// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"2019_2_IBAT/pkg/pkg/auth/session"
	"2019_2_IBAT/pkg/pkg/middleware"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

var addr = flag.String("addr", ":8090", "http service address")

// func serveHome(w http.ResponseWriter, r *http.Request) {
// 	log.Println(r.URL)
// 	if r.URL.Path != "/" {
// 		http.Error(w, "Not found", http.StatusNotFound)
// 		return
// 	}
// 	if r.Method != "GET" {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
// 	http.ServeFile(w, r, "home.html")
// }

func main() {
	flag.Parse()
	router := mux.NewRouter()

	grcpConn, err := grpc.Dial(
		"127.0.0.1:8081",
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("cant connect to grpc")
		// return router, err
	}

	sessManager := session.NewServiceClient(grcpConn)
	// AccessLogOut := new(middleware.AccessLogger)
	// AccessLogOut.StdLogger = log.New(os.Stdout, "STD ", log.LUTC|log.Lshortfile)
	authMiddleware := middleware.AuthMiddlewareGenerator(sessManager)
	fmt.Println(authMiddleware)
	// router.Use(authMiddleware)

	hub := newHub()
	go hub.run()
	// router.HandleFunc("/", serveHome)
	router.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	fmt.Printf("Chat addres %s", *addr)
	err = http.ListenAndServe(*addr, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
