package handlers

import (
	"container/ring"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// we'll use a global channel for the logging goroutine to read from
var (
	log_chan chan string
	// send channels to get data from through this one
	request_chan chan (chan<- []byte)
	requests     uint64
	logs         *ring.Ring
	maxRingSize  int
)

func init() {
	// init the global log_chan and its reader
	log_chan = make(chan string)
	request_chan = make(chan (chan<- []byte))
	maxRingSize = 1000
	logs = ring.New(maxRingSize)
	go loggerRoutine()
}

func MaxLogSize(size int) {
    // set the backlog size (and resets it in the process so call it at the very beginning)
    logs = ring.New(size)
    maxRingSize = size
}

func loggerRoutine() {
	var (
		li       string
		out_chan chan<- []byte
		enc      []byte
		err      error
		log_len  int
	)
	for {
		select {
		case li = <-log_chan:
			log.Printf("LOG \"%s\"", li)
            logs.Value = li
			logs = logs.Next()
			//inc the global counter
			requests += 1
		case out_chan = <-request_chan:
			log.Printf("Logs requested (%d in the ring)\n", logs.Len())
			// get the element in a regular list
			log_len = logs.Len()
			lst := make([]string, log_len)
            var index int = 0
            logs.Do(func(elem interface{}) {
                if elem != nil {
                    //log.Printf("current index is %d\n", index)
                    lst[index] = elem.(string)
                    index += 1
                }
            })
			// reset the ring
			logs = ring.New(maxRingSize)
            enc, err = json.Marshal(lst[0:index])
			if err != nil {
				log.Printf("could not convert the logs to JSON !!!")
			}
			out_chan <- enc
		}
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
