package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	mux.Handle("/", fileServer)
	mux.Handle("/assets/logo.png", fileServer)

	port := "8080"
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Println("Starting server on port 8080...")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error starting server. ERR: %v", err)
	}
}
