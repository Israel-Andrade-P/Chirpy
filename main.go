package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type dtoRequest struct {
	Body string `json:"body"`
}

type dtoResponse struct {
	Valid string `json:"valid"`
}

type errorResponse struct {
	ErrMsg string `json:"error"`
}

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
	html := fmt.Sprintf(`
	        <html>
				<body>
					<h1>Welcome, Chirpy Admin</h1>
					<p>Chirpy has been visited %d times!</p>
				</body>
			</html>
	    `, hits)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
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
	mux.HandleFunc("GET /admin/healthz", readiness)
	mux.HandleFunc("GET /admin/metrics", apicfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apicfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

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

func validateChirp(w http.ResponseWriter, r *http.Request) {
	var reqBody dtoRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error has occurred decoding request body. ERR: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(reqBody.Body) > 140 {
		respErr := errorResponse{ErrMsg: "Chirp is too long"}
		data, err := json.Marshal(respErr)
		if err != nil {
			log.Printf("Error has occurred marshaling data. ERR: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "aplication/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}

	res := dtoResponse{Valid: "true"}
	data, err := json.Marshal(res)
	if err != nil {
		log.Printf("Error has occurred marshaling data. ERR: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
