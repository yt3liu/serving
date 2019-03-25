/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"log"

	"fmt"
	"net/http"
	"sync"
	"time"

	"os"
)

func stopHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Crashed...")
	os.Exit(5)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received a health check request")
	fmt.Fprintf(w, "I'm still healthy")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Hello world received a request.")
	target := os.Getenv("TARGET")
	if target == "" {
		target = "World"
	}
	fmt.Fprintf(w, "Hello %s!\n", target)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		largest := 399989
		msg := fmt.Sprintf("The largest prime under 400000 is %d. Enjoy your noodles!", largest)
		fmt.Fprint(w, msg)
		log.Print(msg)
	}()
	go func() {
		defer wg.Done()
		start := time.Now()
		time.Sleep(time.Second)
		msg := fmt.Sprintf("Slept for %v.", time.Since(start))
		fmt.Fprint(w, msg)
		log.Print(msg)
	}()
	wg.Wait()
}

func main() {
	log.Println("Started")

	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/healthz", healthHandler)
	http.ListenAndServe(":8080", nil)
	//test.ListenAndServeGracefully(":8080", handler)
	//test.ListenAndServeGracefully(":8080", stopHandler)
}
