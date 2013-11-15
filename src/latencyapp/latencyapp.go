package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
    "encoding/json"
)

type Message struct {
    Message string
    Delay int64
}

func handler(w http.ResponseWriter, r *http.Request) {
    var delay int64

	// get the number of seconds to wait to respond
	integerBase := 10
	bitSize := 32
	latency := r.FormValue("delay")
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

    message := Message{"success", delay}
    jsonMessage, err := json.Marshal(message)
    if err != nil {
        fmt.Println("Error marshalling to json!")
        fmt.Println(err)
        jsonMessage = []byte("{\"message\":\"Error marshalling json\"}")
    }

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonMessage[:]))
}

func main() {
	http.HandleFunc("/", handler)              // redirect all urls to the handler function
	http.ListenAndServe("localhost:9999", nil) // listen for connections at port 9999 on the local machine
}
