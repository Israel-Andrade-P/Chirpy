package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Israel-Andrade-P/Chirpy.git/internal/auth"
	"github.com/Israel-Andrade-P/Chirpy.git/internal/database"
	"github.com/Israel-Andrade-P/Chirpy.git/utils"
)

type (
	loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	LoginResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
)

func (cfg *Apiconfig) Login(w http.ResponseWriter, r *http.Request) {
	var loginReq loginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		log.Printf("Error has occurred decoding request body. ERR: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	user, err := cfg.DbQueries.GetUserByEmail(r.Context(), loginReq.Email)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Email")
		return
	}
	isMatch, err := auth.CheckPasswordHash(loginReq.Password, user.Password)
	if err != nil {
		log.Printf("Error has occurred comparing hashed password. ERR: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !isMatch {
		utils.RespondWithError(w, http.StatusUnauthorized, "email or password incorrect.")
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.Secret, time.Minute*time.Duration(cfg.Expiration))
	if err != nil {
		log.Printf("Error has occurred creating the jwt. ERR: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	refToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error has occurred creating the refersh token. ERR: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	rToken, err := cfg.DbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour)})
	if err != nil {
		log.Printf("DB error has occurred: %v", err)
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid token")
		return
	}
	loginRes := LoginResponse{AccessToken: jwt, RefreshToken: rToken}
	utils.RespondWithJson(w, http.StatusOK, loginRes)
}
