package main

import (
	"log"
	"net/http"
    "logserver/handlers"
)

func main() {
    // handlers
    http.HandleFunc("/log", handlers.HandleLogPost)
    http.HandleFunc("/stats", handlers.HandleStats)

    err := http.ListenAndServe(":8888", nil)
    if err != nil {
        log.Fatal("while listening:", err)
    }
}
