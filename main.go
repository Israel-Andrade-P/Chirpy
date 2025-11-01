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
	"time"

	"github.com/Israel-Andrade-P/Chirpy.git/internal/database"
	"github.com/google/uuid"
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

type registerRequest struct {
	Email string `json:"email"`
}

type ResponseUser struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type (
	chirpRequest struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	chirpResponse struct {
		ID        uuid.UUID `json:"id"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		UserID    uuid.UUID `json:"user_id"`
	}
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
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

func (cfg *apiConfig) registerUser(w http.ResponseWriter, r *http.Request) {
	var registerReq registerRequest
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		log.Printf("Error has occurred decoding request body. ERR: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Error")
		return
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), registerReq.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Email")
	}
	u := ResponseUser{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
	respondWithJson(w, http.StatusCreated, u)
}

func (cfg *apiConfig) saveChirp(w http.ResponseWriter, r *http.Request) {
	var chirpReq chirpRequest
	if err := json.NewDecoder(r.Body).Decode(&chirpReq); err != nil {
		log.Printf("Error has occurred decoding request body. ERR: %v\n", err)
		respondWithError(w, http.StatusInternalServerError, "Internal Error")
		return
	}
	if len(chirpReq.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long.")
		return
	}
	cleanMessage := cleanUpMessage(chirpReq.Body)
	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleanMessage, UserID: chirpReq.UserID})
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Chirp")
	}
	c := chirpResponse{ID: chirp.ID, Body: chirp.Body, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, UserID: chirp.UserID}
	respondWithJson(w, http.StatusCreated, c)
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("You are not allowed to do this"))
		return
	}
	cfg.fileserverHits.Store(0)
	if err := cfg.dbQueries.DeleteUsers(r.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal Error")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Counter reset.\nAlso all your users are gone btw\n"))
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("ERROR >> %v", err)
	}

	dbUrl := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("ERROR >> %v", err)
	}
	dbQueries := database.New(db)

	apicfg := &apiConfig{dbQueries: dbQueries, platform: platform}
	mux := http.NewServeMux()

	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	fileServer = noCacheFileServer(fileServer)

	mux.Handle("/app/", apicfg.middlewareMetricsInc(fileServer))
	mux.Handle("/app/assets/logo.png", apicfg.middlewareMetricsInc(fileServer))
	mux.HandleFunc("GET /admin/healthz", readiness)
	mux.HandleFunc("GET /admin/metrics", apicfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apicfg.handlerReset)
	mux.HandleFunc("POST /api/users", apicfg.registerUser)
	mux.HandleFunc("POST /api/chirps", apicfg.saveChirp)

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
