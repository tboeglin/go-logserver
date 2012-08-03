package main

import (
	"flag"
	"fmt"
	"log"
	"logserver/handlers"
	"net/http"
)

func main() {
	var (
		portno  int
		backlogsize int
	)
	flag.IntVar(&portno, "port", 8888, "the port to bind to")
	flag.IntVar(&backlogsize, "backlogsize", 1000, "the size of the backlog to keep in memory for fetching over HTTP")

	flag.Parse()
	handlers.MaxLogSize(backlogsize)
	// handlers
	http.HandleFunc("/log", handlers.HandleLogPost)
	http.HandleFunc("/stats", handlers.HandleStats)

    log.Printf("Starting on port %d with backlog size %d\n", portno, backlogsize)

	err := http.ListenAndServe(fmt.Sprintf(":%d", portno), nil)
	if err != nil {
		log.Fatal("while listening:", err)
	}
}
