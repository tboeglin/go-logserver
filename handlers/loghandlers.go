package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

type LogInfo struct {
	Source   string
	Severity string
	Message  string
}

// we'll use a global channel for the logging goroutine to read from
var (
	log_chan chan LogInfo
	requests uint64
	once     sync.Once
)

func logFromChan() {
	var li LogInfo
	for {
		li = <-log_chan
		log.Printf("LOG [%s] from %s: %s", li.Severity, li.Source, li.Message)
		//inc the global counter
		requests += 1
	}
}

func initModule() {
	// init the global log_chan and its reader
	log_chan = make(chan LogInfo)
	go logFromChan()
}

func Init() {
	once.Do(initModule)
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
	var li LogInfo
	err = json.Unmarshal(payload, &li)
	if err != nil {
		log.Printf("could not decode json \"%s\"", payload)
        rw.WriteHeader(http.StatusBadRequest)
		return
	}
	// send it to the logger
	log_chan <- li
    rw.WriteHeader(http.StatusOK)
}
