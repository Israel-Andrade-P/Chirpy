package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/Israel-Andrade-P/Chirpy.git/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
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
	if err := godotenv.Load(); err != nil {
		log.Fatalf("ERROR >> %v", err)
	}

	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("ERROR >> %v", err)
	}
	dbQueries := database.New(db)
	log.Println(dbQueries)

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
	if err := server.ListenAndServe(); err != nil {
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
		respondWithError(w, http.StatusInternalServerError, "Internal Error")
		return
	}

	if len(reqBody.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long.")
		return
	}

	cleanMessage := cleanUpMessage(reqBody.Body)
	res := dtoResponse{Valid: cleanMessage}
	respondWithJson(w, http.StatusOK, res)
}

func respondWithJson(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error has occurred marshaling data. ERR: %v\n", err)
		//Fallback response
		http.Error(w, `"error": "Internal server error"`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	w.Write(data)
}

func respondWithError(w http.ResponseWriter, status int, errMsg string) {
	errPayload := errorResponse{ErrMsg: errMsg}
	respondWithJson(w, status, errPayload)
}

func cleanUpMessage(message string) string {
	profanityList := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(strings.ToLower(message), " ")

	for i := 0; i < len(words); i++ {
		if contains(profanityList, words[i]) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func contains[T comparable](slice []T, target T) bool {
	for _, element := range slice {
		if element == target {
			return true
		}
	}
	return false
}
