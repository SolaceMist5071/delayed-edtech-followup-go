package main

import (
	"log"
	"net/http"
)

func main() {
	client, err := NewInfraiClient()
	if err != nil {
		log.Fatal(err)
	}
	service := &Service{client: client}
	mux := http.NewServeMux()
	mux.HandleFunc("/followups", service.schedule)
	mux.HandleFunc("/report", service.report)
	log.Println("edtech follow-up service listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
