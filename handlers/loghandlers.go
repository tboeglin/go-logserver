package handlers

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"container/ring"
)

// we'll use a global channel for the logging goroutine to read from
var (
	log_chan chan string
    // send channels to get data from through this one
    request_chan <-chan (chan<- []string)
	requests uint64
    logs *ring.Ring
)

func init() {
	// init the global log_chan and its reader
	log_chan = make(chan string)
    request_chan = make(<-chan (chan<- []string))
    logs = ring.New(1000)
	go loggerRoutine()
}

func loggerRoutine() {
	var li string
	for {
		li = <-log_chan
        log.Printf("LOG \"%s\"", li)
		//inc the global counter
		requests += 1
	}
}

func HandleStats(rw http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        log.Printf("HandleStats: got a non get request from %s\n", r.RemoteAddr)
        return
    }
    fmt.Fprintln(rw, "logged up to now:", requests)
}

func HandleLogPost(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
        log.Printf("handleLogPost: got an non-POST request from %s\n", r.RemoteAddr)
        rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//try decoding the payload as json
	var (
		payload []byte
		err     error
	)
	payload, err = ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("could not get body from request?!")
        rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	// send it to the logger
	log_chan <- string(payload)
    rw.WriteHeader(http.StatusOK)
}
