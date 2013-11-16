package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	rsp, err := http.Get("http://localhost:9000/sample")
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(rsp.Body)

	var wg sync.WaitGroup

	var numResults int32
	for scanner.Scan() {
		url := scanner.Text()
		wg.Add(1)
		go func(aURL string) {
			defer wg.Done()
			rsp, err := http.Get(aURL)
			if err != nil {
				log.Print(err)
				return
			}
			io.Copy(os.Stdout, rsp.Body)
			fmt.Println()
			atomic.AddInt32(&numResults, 1)
		}(url)
	}

	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		log.Printf("Got all results")
	case <-time.After(6 * time.Second):
		log.Printf("Timed out")
	}

	log.Printf("Got %d results", numResults)
}
