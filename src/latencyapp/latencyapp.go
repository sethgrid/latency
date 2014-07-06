package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/nu7hatch/gouuid"
)

// contains the payload of a single request
type Message struct {
	Message    string `json:"message"`
	Delay      int64  `json:"delay"`
	StatusCode int    `json:"status_code"`
}

// each url in the sample page
type Url struct {
	Url string `json:"url"`
}

// container for each Url in the sample page
type UrlList struct {
	Urls []Url `json:"urls"`
}

// marshler and message Marshler act as a type of middle ware to reduce code duplication
type marshler func(v UrlList) (string, string)
type messageMarshler func(v Message) (string, string)

// used to find out what function was passed into the marshler func
// allows introspection and to keep the right kind of links (txt,json,xml)
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Sample Handler displays the sample page with all the urls to attempt
func jsonSampleHandler(w http.ResponseWriter, r *http.Request) {
	dynamicSampleHandler(jsonData, w, r)
}

// Sample Handler displays the sample page with all the urls to attempt
func xmlSampleHandler(w http.ResponseWriter, r *http.Request) {
	dynamicSampleHandler(xmlData, w, r)
}

// Sample Handler displays the sample page with all the urls to attempt
func txtSampleHandler(w http.ResponseWriter, r *http.Request) {
	dynamicSampleHandler(txtData, w, r)
}

// passed in to make content json
func jsonData(v UrlList) (string, string) {
	res, _ := json.Marshal(v)
	return "application/json", string(res)
}
func jsonMessageData(v Message) (string, string) {
	message, err := json.Marshal(v)
	if err != nil {
		fmt.Println("Error marshalling to json!")
		fmt.Println(err)
		message = []byte("{\"message\":\"Error marshalling json\"}")
	}
	return "application/json", string(message)
}

// passed in to make data xml
func xmlData(v UrlList) (string, string) {
	res, _ := xml.Marshal(v)
	return "application/xml", string(res)
}
func xmlMessageData(v Message) (string, string) {
	message, err := xml.Marshal(v)
	if err != nil {
		fmt.Println("Error marshalling to xml!")
		fmt.Println(err)
		message = []byte("{\"message\":\"Error marshalling xml\"}")
	}
	return "application/xml", string(message)
}

// passed in to make data text
func txtData(v UrlList) (string, string) {
	res := ""
	for _, url := range v.Urls {
		res += fmt.Sprintf("%s\n", url)
	}
	res = strings.Replace(res, "{", "", -1)
	res = strings.Replace(res, "}", "", -1)
	return "text/plain", res
}
func txtMessageData(v Message) (string, string) {
	message := fmt.Sprintf("message: %s, delay: %d\n", v.Message, v.Delay)
	return "text/plain", message
}

func DataFromRequest(r *http.Request) Message {
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

	message := Message{Delay: delay}

	codeStr := r.URL.Query().Get("code")

	// Use the code we got passed in
	if err == nil && codeStr != "" {
		message.StatusCode, err = strconv.Atoi(codeStr)
		if err != nil {
			message.StatusCode = 200
		}

		// Pick a random message.StatusCode
	} else {
		rnd := rand.Intn(10)
		if rnd == 4 {
			message.StatusCode = 400
		} else if rnd == 5 {
			message.StatusCode = 500
		} else {
			message.Message = "success"
			message.StatusCode = 200
		}
	}

	if message.StatusCode >= 500 {
		message.Message = "server error"
	} else if message.StatusCode >= 400 {
		message.Message = "client error"
	} else {
		message.Message = "success"
	}

	return message
}

type DynamicHandler struct {
	marshaler messageMarshler
}

// takes a message and uses the messageMarshler to set the type of message
func (d *DynamicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	message := DataFromRequest(r)

	log.Printf("Going to wait %d seconds...\n", message.Delay)
	time.Sleep(time.Duration(message.Delay) * time.Second)

	//m = txtMessageData
	contentType, msg := d.marshaler(message)

	w.Header().Set("Content-Type", contentType)
	log.Printf("Sending status code %d", message.StatusCode)
	w.WriteHeader(message.StatusCode)

	fmt.Fprintf(w, "%s", msg)
}

// takes the marshler to determine the type of sample page to display
func dynamicSampleHandler(m marshler, w http.ResponseWriter, r *http.Request) {

	numUrlsString := r.URL.Query().Get("n")
	numUrls := 100
	if numUrlsString != "" {
		numUrls, _ = strconv.Atoi(numUrlsString)
	}

	urlType := ""
	mName := GetFunctionName(m)
	log.Print("m's name is ", mName)
	if strings.Contains(mName, "json") {
		urlType = "json"
	} else if strings.Contains(mName, "xml") {
		urlType = "xml"
	} else if strings.Contains(mName, "txt") {
		urlType = "txt"
	} else {
		urlType = "json"
		log.Print("defaulting url to json. unknown url type in function name ", mName)
	}

	urlList := UrlList{}
	for i := 0; i < numUrls; i++ {
		u, err := uuid.NewV4()
		if err != nil {
			fmt.Println("Error with uuid")
			u = nil
		}
		url := Url{Url: fmt.Sprintf("http://%s/%s/%s", r.Host, urlType, u)}
		urlList.Urls = append(urlList.Urls, url)
	}
	contentType, message := m(urlList)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", message)
}

func main() {
	// set up handlers to serve up json, xml, and plain text
	jsonHandler := &DynamicHandler{marshaler: jsonMessageData}
	http.Handle("/json/", jsonHandler)
	http.HandleFunc("/json/sample", jsonSampleHandler)

	xmlHandler := &DynamicHandler{marshaler: xmlMessageData}
	http.Handle("/xml/", xmlHandler)
	http.HandleFunc("/xml/sample", xmlSampleHandler)

	txtHandler := &DynamicHandler{marshaler: txtMessageData}
	http.Handle("/txt/", txtHandler)
	http.HandleFunc("/txt/sample", txtSampleHandler)

	// default (anything not matching above will fall to the jsonHandler)
	http.Handle("/", jsonHandler)
	http.HandleFunc("/sample", jsonSampleHandler)

	// serve
	log.Printf("Listening on ':%s'", os.Getenv("PORT"))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil) // listen for connections at port 9999 on the local machine
	if err != nil {
		log.Printf("Failed to start server: %s", err)
	}
}
