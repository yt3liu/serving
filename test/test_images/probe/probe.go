package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	started := time.Now()

	http.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received healthy request...")
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})

	http.HandleFunc("/unhealthy", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received unhealthy request...")
		w.WriteHeader(500)
		w.Write([]byte("not ok"))
	})

	// Reference: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/#define-a-liveness-command
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received healthz request...")
		duration := time.Since(started)
		if duration.Seconds() > 10 {
			log.Println("healthz reported 500")
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("error: %v", duration.Seconds())))
		} else {
			log.Println("healthz reported 200")
			w.WriteHeader(200)
			w.Write([]byte("ok"))
		}
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}
