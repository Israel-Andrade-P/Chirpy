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
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	UserResponse struct {
		ID        uuid.UUID `json:"id"`
		Email     string    `json:"email"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Token     string    `json:"token"`
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
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Email")
		return
	}
	u := UserResponse{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
	utils.RespondWithJson(w, http.StatusCreated, u)
}

func (cfg *Apiconfig) Login(w http.ResponseWriter, r *http.Request) {
	var userReq userRequest
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		log.Printf("Error has occurred decoding request body. ERR: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	user, err := cfg.DbQueries.GetUserByEmail(r.Context(), userReq.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Email")
		return
	}
	isMatch, err := auth.CheckPasswordHash(userReq.Password, user.Password)
	if err != nil {
		log.Printf("Error has occurred comparing hashed password. ERR: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !isMatch {
		utils.RespondWithError(w, http.StatusUnauthorized, "email or password incorrect.")
		return
	}
	expiration := 60
	if userReq.ExpiresInSeconds > 0 && userReq.ExpiresInSeconds <= expiration {
		expiration = userReq.ExpiresInSeconds
	}
	jwt, err := auth.MakeJWT(user.ID, cfg.Secret, time.Minute*time.Duration(expiration))
	if err != nil {
		log.Printf("Error has occurred creating the jwt. ERR: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	userRes := UserResponse{ID: user.ID, Email: user.Email, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt, Token: jwt}
	utils.RespondWithJson(w, http.StatusOK, userRes)
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
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Counter reset.\nAlso all your users are gone btw\n"))
}
