package main

import (
	"bytes"
	"encoding/base64"
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

// Marshaler and message Marshaler act as a type of middle ware to reduce code duplication
type Marshaler func(v UrlList) (string, string)
type messageMarshaler func(v Message) (string, string)

// decodes urls and returns the values for status, message, and delay
type decoder func(*http.Request) Message

// generates an encoded url
type encoder func() string

// used to find out what function was passed into the Marshaler func
// allows introspection and to keep the right kind of links (txt,json,xml)
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// passed in to make content json
func jsonData(v UrlList) (string, string) {
	res, _ := json.Marshal(v)
	return "application/json", string(res)
}
func jsonMessageData(v Message) (string, string) {
	message, err := json.Marshal(v)
	if err != nil {
		log.Println("Error marshalling to json!")
		log.Println(err)
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
		log.Println("Error marshalling to xml!")
		log.Println(err)
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

// decoder that ignores the request and returns random data
func RandomDecoder(r *http.Request) Message {
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

func Base64JSONDecoder(r *http.Request) Message {
	// grab the last piece of the url
	urlParts := strings.Split(r.URL.Path, "/")
	msgBase64JSON := urlParts[len(urlParts)-1]
	reader := strings.NewReader(msgBase64JSON)
	b64Decoder := base64.NewDecoder(base64.URLEncoding, reader)
	jsonDecoder := json.NewDecoder(b64Decoder)

	message := Message{}
	jsonDecoder.Decode(&message)

	return message
}

type DelayHandler struct {
	marshaler messageMarshaler
	decoder   decoder
}

// takes a message and uses the messageMarshaler to set the type of message
func (d *DelayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	message := d.decoder(r)

	log.Printf("Going to wait %d seconds...", message.Delay)
	time.Sleep(time.Duration(message.Delay) * time.Second)

	//m = txtMessageData
	contentType, msg := d.marshaler(message)

	w.Header().Set("Content-Type", contentType)
	log.Printf("Sending status code %d", message.StatusCode)
	w.WriteHeader(message.StatusCode)

	fmt.Fprintf(w, "%s", msg)
}

func UUIDEncoder() string {
	u, err := uuid.NewV4()
	if err != nil {
		log.Println("Error with uuid")
		return ""
	}

	return u.String()
}

func RandomEncoder() string {
	message := Message{}

	// random delay
	message.Delay = rand.Int63n(10)

	// random status code
	// determine message here too
	rndCode := rand.Intn(10)
	switch rndCode {
	case 4:
		message.StatusCode = 400
		message.Message = "client error"
	case 5:
		message.StatusCode = 500
		message.Message = "server error"
	default:
		message.StatusCode = 200
		message.Message = "success"
	}

	jsonBytes, err := json.Marshal(message)
	if err != nil {
		log.Println("Error with uuid")
		return "broken-url"
	}

	// buffer for storing the result of the base64 encoding
	buffer := &bytes.Buffer{}

	// build encoder with URLEncoding for safe url usage
	encoder := base64.NewEncoder(base64.URLEncoding, buffer)
	encoder.Write(jsonBytes)
	// close required to flush remaining bytes
	encoder.Close()

	return string(buffer.Bytes())
}

type SampleHandler struct {
	marshaler Marshaler
	encoder   encoder

	urlType string
}

// takes the Marshaler to determine the type of sample page to display
func (s *SampleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	numUrlsString := r.URL.Query().Get("n")
	numUrls := 100
	if numUrlsString != "" {
		numUrls, _ = strconv.Atoi(numUrlsString)
	}

	urlList := UrlList{}
	for i := 0; i < numUrls; i++ {
		urlData := s.encoder()
		url := Url{Url: fmt.Sprintf("http://%s/%s/%s", r.Host, s.urlType, urlData)}
		urlList.Urls = append(urlList.Urls, url)
	}
	contentType, message := s.marshaler(urlList)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s", message)
}

func main() {
	// set up handlers to serve up json, xml, and plain text
	jsonHandler := &DelayHandler{marshaler: jsonMessageData, decoder: Base64JSONDecoder}
	http.Handle("/json/", jsonHandler)

	jsonSampleHandler := &SampleHandler{marshaler: jsonData, urlType: "json", encoder: RandomEncoder}
	http.Handle("/json/sample", jsonSampleHandler)

	xmlHandler := &DelayHandler{marshaler: xmlMessageData, decoder: Base64JSONDecoder}
	http.Handle("/xml/", xmlHandler)

	xmlSampleHandler := &SampleHandler{marshaler: xmlData, urlType: "xml", encoder: RandomEncoder}
	http.Handle("/xml/sample", xmlSampleHandler)

	txtHandler := &DelayHandler{marshaler: txtMessageData, decoder: Base64JSONDecoder}
	http.Handle("/txt/", txtHandler)

	txtSampleHandler := &SampleHandler{marshaler: txtData, urlType: "txt", encoder: RandomEncoder}
	http.Handle("/txt/sample", txtSampleHandler)

	// default (anything not matching above will fall to the jsonHandler)
	http.Handle("/", jsonHandler)
	http.Handle("/sample", jsonSampleHandler)

	// serve
	log.Printf("Listening on ':%s'", os.Getenv("PORT"))
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil) // listen for connections at port 9999 on the local machine
	if err != nil {
		log.Printf("Failed to start server: %s", err)
	}
}
