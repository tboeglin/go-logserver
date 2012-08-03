package handlers

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type LogInfo struct {
	Timestamp time.Time // will be marshalled according to RFC3339
	Severity  string
	Source    string
	Message   string
}

// we'll use a global channel for the logging goroutine to read from
var (
	log_chan chan LogInfo
	// send channels to get data from through this one
	request_chan chan (chan<- []byte)
	requests     uint64
	logs         *ring.Ring
	maxRingSize  int
)

func init() {
	// init the global log_chan and its reader
	log_chan = make(chan LogInfo)
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
		li       LogInfo
		out_chan chan<- []byte
		enc      []byte
		err      error
		log_len  int
	)
	for {
		select {
		case li = <-log_chan:
			log.Printf("[%s] from %s: %s\n", li.Severity, li.Source, li.Message)
			logs.Value = li
			logs = logs.Next()
			//inc the global counter
			requests += 1
		case out_chan = <-request_chan:
			// get the element in a regular list
			log_len = logs.Len()
			lst := make([]LogInfo, log_len)
			var index int = 0
			logs.Do(func(elem interface{}) {
				if elem != nil {
					//log.Printf("current index is %d\n", index)
					lst[index] = elem.(LogInfo)
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
	switch r.Method {
	case "POST":
		var (
			payload []byte
			err     error
			li      LogInfo
		)
		payload, err = ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("could not get body from request?!")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		//try decoding the payload as json
		err = json.Unmarshal(payload, &li)
		if err != nil {
			log.Printf("could not unmarshal JSON \"%s\" from %s\n",
				payload, r.RemoteAddr)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		// add the timestamp
		li.Timestamp = time.Now().UTC()
		// send it to the logger
		log_chan <- li
		rw.WriteHeader(http.StatusOK)
	case "GET":
		log.Printf("logs requested from %s\n", r.RemoteAddr)
		var logs_json []byte
		logs_chan := make(chan []byte)
		request_chan <- logs_chan
		logs_json = <-logs_chan
		rw.Header().Add("Content-Type", "application/json")
		_, err := rw.Write(logs_json)
		if err != nil {
			log.Printf("Could not send back json: %s\n", err.Error())
		}

	default:
		log.Printf("handleLogPost: got an non POST/GET request from %s\n", r.RemoteAddr)
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}
