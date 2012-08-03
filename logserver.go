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
		logsize int
	)
	flag.IntVar(&portno, "port", 8888, "the port to bind to")
	flag.IntVar(&logsize, "logsize", 1000, "the size of the backlog")

	flag.Parse()
	handlers.MaxLogSize(logsize)
	// handlers
	http.HandleFunc("/log", handlers.HandleLogPost)
	http.HandleFunc("/stats", handlers.HandleStats)

    log.Printf("Starting on port %d with backlog size %d\n", portno, logsize)

	err := http.ListenAndServe(fmt.Sprintf(":%d", portno), nil)
	if err != nil {
		log.Fatal("while listening:", err)
	}
}
