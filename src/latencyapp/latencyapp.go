package main

import (
	"encoding/json"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Message struct {
	Message string `json:"message"`
	Delay   int64  `json:"delay"`
}

type Url struct {
	Url string `json:"url"`
}

type UrlList struct {
	Urls []Url `json:"urls"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	var delay int64

	// get the number of seconds to wait to respond
	latency := r.FormValue("delay")
	seconds, err := strconv.Atoi(latency)
	if err != nil {
		seconds = 0
	}
	if seconds == 0 {
		delay = rand.Int63n(10)
	} else {
		delay = int64(seconds)
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
	fmt.Fprintf(w, "%s", jsonMessage)
}

func sampleHandler(w http.ResponseWriter, r *http.Request) {
	numUrlsString := r.URL.Query().Get("n")
	numUrls := 100
	if numUrlsString != "" {
		numUrls, _ = strconv.Atoi(numUrlsString)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	urlList := UrlList{}
	for i := 0; i < numUrls; i++ {
		u, err := uuid.NewV4()
		if err != nil {
			fmt.Println("Error with uuid")
			u = nil
		}
		url := Url{Url: fmt.Sprintf("http://%s/%s", r.Host, u)}
		urlList.Urls = append(urlList.Urls, url)
	}
	jsonMessage, _ := json.Marshal(urlList)
	fmt.Fprintf(w, "%s", jsonMessage)
}

func main() {
	http.HandleFunc("/", handler)             // redirect all urls to the handler function
	http.HandleFunc("/sample", sampleHandler) // get a list of URLs

	log.Printf("Listening on ':%s'", os.Getenv("PORT"))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil) // listen for connections at port 9999 on the local machine
	if err != nil {
		log.Printf("Failed to start server: %s", err)
	}
}
