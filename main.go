package main

import (
	"2019_2_IBAT/pkg/pkg/auth/session"
	"2019_2_IBAT/pkg/pkg/middleware"

	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

var addr = flag.String("addr", ":8090", "http service address")

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
	authMiddleware := middleware.AuthMiddlewareGenerator(sessManager)

	fmt.Println(authMiddleware)
	router.Use(authMiddleware)

	hub := newHub()
	go hub.run()
	router.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	fmt.Printf("Chat addres %s", *addr)
	err = http.ListenAndServe(*addr, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
