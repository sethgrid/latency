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
	Path    string `json:"path"`
}

type Url struct {
	Url string `json:"url"`
}

type UrlList struct {
	Urls []Url `json:"urls"`
}

// Takes `code`
func handler(w http.ResponseWriter, r *http.Request) {
	var delay int64

	log.Printf("%s %s", r.Method, r.URL.String())

	// get the number of seconds to wait to respond
	latency := r.URL.Query().Get("delay")
	seconds, err := strconv.Atoi(latency)

	if err != nil || seconds < 0 {
		delay = rand.Int63n(10)
	} else {
		delay = int64(seconds)
	}

	message := Message{Delay: delay, Path: r.URL.Path}

	log.Printf("Going to wait %d seconds...\n", delay)
	time.Sleep(time.Duration(delay) * time.Second)

	var code int
	var rnd int
	codeStr := r.URL.Query().Get("code")

	// Use the code we got passed in
	if err == nil && codeStr != "" {
		code, err = strconv.Atoi(codeStr)
		if err != nil {
			code = 200
		}

		// Pick a random code
	} else {
		rnd = rand.Intn(10)
		if rnd == 4 {
			code = 400
		} else if rnd == 5 {
			code = 500
		} else {
			message.Message = "success"
			code = 200
		}
	}

	if code >= 500 {
		message.Message = "server error"
	} else if code >= 400 {
		message.Message = "client error"
	} else {
		message.Message = "success"
	}
	log.Printf("Sending status code %d", code)
	w.WriteHeader(code)

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
