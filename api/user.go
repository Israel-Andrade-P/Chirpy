package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Israel-Andrade-P/Chirpy.git/internal/auth"
	"github.com/Israel-Andrade-P/Chirpy.git/internal/database"
	"github.com/Israel-Andrade-P/Chirpy.git/utils"
	"github.com/google/uuid"
)

type (
	userRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	UserResponse struct {
		ID        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

func (cfg *Apiconfig) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var registerReq userRequest
	if err := json.NewDecoder(r.Body).Decode(&registerReq); err != nil {
		log.Printf("Error has occurred decoding request body. ERR: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	hashedPwd, err := auth.HashPassword(registerReq.Password)
	if err != nil {
		log.Printf("Error has occurred hashing password. ERR: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}

	user, err := cfg.DbQueries.CreateUser(r.Context(), database.CreateUserParams{Email: registerReq.Email, Password: hashedPwd})
	if err != nil {
		log.Printf("DB error has occurred: %v", err)
		utils.RespondWithError(w, http.StatusBadRequest, "invalid email")
		return
	}
	u := UserResponse{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
	utils.RespondWithJson(w, http.StatusCreated, u)
}

func (cfg *Apiconfig) DeleteAllUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("you are not allowed to do this"))
		return
	}
	cfg.FileserverHits.Store(0)
	if err := cfg.DbQueries.DeleteUsers(r.Context()); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Counter reset.\nAlso all your users are gone btw\n"))
}

func (cfg *Apiconfig) UpdateUser(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "please autenticate yourself")
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "your token is invalid, get a new one")
		return
	}
	var req userRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error has occurred decoding request body. ERR: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	hashedPwd, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Error has occurred hashing password. ERR: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}

	updatedUser, err := cfg.DbQueries.UpdateUser(r.Context(), database.UpdateUserParams{Email: req.Email, Password: hashedPwd, ID: userId})
	if err != nil {
		log.Printf("DB error has occurred: %v", err)
		utils.RespondWithError(w, http.StatusBadRequest, "invalid email")
		return
	}
	utils.RespondWithJson(w, http.StatusOK, UserResponse{
		ID:        updatedUser.ID,
		Email:     updatedUser.Email,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	})
}
