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

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("Starting server on port 8080...")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error starting server. ERR: %v", err)
	}
}
