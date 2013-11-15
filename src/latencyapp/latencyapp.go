package main

import (
	"fmt"
	"math/rand"
	"net/http" //package for http based web programs
	"strconv"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
    var delay int64

	// get the number of seconds to wait to respond
	integerBase := 10
	bitSize := 32
	latency := r.FormValue("latency")
	seconds, err := strconv.ParseInt(latency, integerBase, bitSize)
    if err != nil {
        seconds = 0
    }
    if seconds == 0  {
		delay = rand.Int63n(10)
	} else {
        delay = seconds
    }

	fmt.Printf("Going to wait %d seconds...\n", delay)

    time.Sleep(time.Duration(delay) * time.Second)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"message\":\"success\"}")
}

func main() {
	http.HandleFunc("/", handler)              // redirect all urls to the handler function
	http.ListenAndServe("localhost:9999", nil) // listen for connections at port 9999 on the local machine
}
