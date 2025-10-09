package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))

	mux.Handle("/app/", fileServer)
	mux.Handle("/app/assets/logo.png", fileServer)
	mux.HandleFunc("/healthz", readiness)

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

func readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
