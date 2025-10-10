package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	hits := cfg.fileserverHits.Load()
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "Hits: %d", hits)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Counter reset.\n"))
}

func main() {
	apicfg := &apiConfig{}
	mux := http.NewServeMux()

	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	fileServer = noCacheFileServer(fileServer)

	mux.Handle("/app/", apicfg.middlewareMetricsInc(fileServer))
	mux.Handle("/app/assets/logo.png", apicfg.middlewareMetricsInc(fileServer))
	mux.HandleFunc("GET /healthz", readiness)
	mux.HandleFunc("GET /metrics", apicfg.handlerMetrics)
	mux.HandleFunc("POST /reset", apicfg.handlerReset)

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

func noCacheFileServer(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("If-Modified-Since")
		r.Header.Del("If-None-Match")
		h.ServeHTTP(w, r)
	})
}
