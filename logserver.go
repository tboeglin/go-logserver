package main

import (
	"flag"
	"fmt"
	"log"
	"github.com/tboeglin/go-logserver/handlers"
    "github.com/tboeglin/go-loggers/rotatingfile"
	"net/http"
)

func main() {
	var (
		portno  int
		backlogsize int
        logfile string
        logsize int
	)
	flag.IntVar(&portno, "port", 8888, "the port to bind to")
	flag.IntVar(&backlogsize, "backlogsize", 1000, "the size of the backlog to keep in memory for fetching over HTTP")
    flag.StringVar(&logfile, "logfile", "STDERR", "the file to log to, by default the console on stderr")
	flag.IntVar(&logsize, "logsize", 10485760, "the maximum size of the log file before it's rotated")

	flag.Parse()
	handlers.MaxLogSize(backlogsize)
    if logfile != "STDERR" {
        log.SetOutput(rotatingfile.Create(logfile, logsize, 10))
    }
	// handlers
	http.HandleFunc("/log", handlers.HandleLogPost)
	http.HandleFunc("/stats", handlers.HandleStats)

    log.Printf("Starting on port %d with backlog size %d\n", portno, backlogsize)

	err := http.ListenAndServe(fmt.Sprintf(":%d", portno), nil)
	if err != nil {
		log.Fatal("while listening:", err)
	}
}
