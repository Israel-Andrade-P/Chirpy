package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Israel-Andrade-P/Chirpy.git/internal/database"
	"github.com/Israel-Andrade-P/Chirpy.git/utils"
	"github.com/google/uuid"
)

type (
	userRequest struct {
		Email string `json:"email"`
	}
	UserResponse struct {
		ID        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

type Apiconfig struct {
	FileserverHits atomic.Int32
	DbQueries      *database.Queries
	Platform       string
}

func (cfg *Apiconfig) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var registerReq userRequest
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		log.Printf("Error has occurred decoding request body. ERR: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Error")
		return
	}
	user, err := cfg.DbQueries.CreateUser(r.Context(), registerReq.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Email")
	}
	u := UserResponse{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
	utils.RespondWithJson(w, http.StatusCreated, u)
}

func (cfg *Apiconfig) DeleteAllUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("You are not allowed to do this"))
		return
	}
	cfg.FileserverHits.Store(0)
	if err := cfg.DbQueries.DeleteUsers(r.Context()); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Error")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Counter reset.\nAlso all your users are gone btw\n"))
}
