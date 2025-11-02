package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Israel-Andrade-P/Chirpy.git/api"
	"github.com/Israel-Andrade-P/Chirpy.git/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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

	apicfg := &api.Apiconfig{DbQueries: dbQueries, Platform: platform}
	mux := http.NewServeMux()

	fileServer := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	fileServer = api.NoCacheFileServer(fileServer)

	mux.Handle("/app/", apicfg.MiddlewareMetricsInc(fileServer))
	mux.Handle("/app/assets/logo.png", apicfg.MiddlewareMetricsInc(fileServer))
	mux.HandleFunc("GET /admin/healthz", api.Readiness)
	mux.HandleFunc("GET /admin/metrics", apicfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", apicfg.DeleteAllUsers)
	mux.HandleFunc("POST /api/users", apicfg.RegisterUser)
	mux.HandleFunc("POST /api/login", apicfg.Login)
	mux.HandleFunc("POST /api/chirps", apicfg.SaveChirp)
	mux.HandleFunc("GET /api/chirps", apicfg.GetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apicfg.GetChirp)

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
