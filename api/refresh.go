package api

import (
	"log"
	"net/http"
	"time"

	"github.com/Israel-Andrade-P/Chirpy.git/internal/auth"
	"github.com/Israel-Andrade-P/Chirpy.git/utils"
)

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}

func (cfg *Apiconfig) Refresh(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "please autenticate yourself")
		return
	}
	dbToken, err := cfg.DbQueries.GetRefreshToken(r.Context(), rToken)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "no token, please log in")
		return
	}
	if dbToken.ExpiresAt.Before(time.Now()) {
		utils.RespondWithError(w, http.StatusUnauthorized, "expired token, please log in")
		return
	}
	if dbToken.RevokedAt.Valid {
		log.Printf("token revoked at: %v", dbToken.RevokedAt.Time)
		utils.RespondWithError(w, http.StatusUnauthorized, "token revoked")
		return
	}

	aToken, err := auth.MakeJWT(dbToken.UserID, cfg.Secret, time.Minute*time.Duration(cfg.Expiration))
	if err != nil {
		log.Printf("Error has occurred creating the jwt. ERR: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	utils.RespondWithJson(w, http.StatusOK, RefreshResponse{AccessToken: aToken})
}

func (cfg *Apiconfig) RevokeToken(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "please autenticate yourself")
		return
	}
	cfg.DbQueries.RevokeToken(r.Context(), rToken)
	w.WriteHeader(http.StatusNoContent)
}
