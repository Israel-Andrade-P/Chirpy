package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Israel-Andrade-P/Chirpy.git/internal/auth"
	"github.com/Israel-Andrade-P/Chirpy.git/utils"
	"github.com/google/uuid"
)

type (
	webHookRequest struct {
		Event string   `json:"event"`
		Data  userInfo `json:"data"`
	}
	userInfo struct {
		UserId string `json:"user_id"`
	}
)

func (cfg *Apiconfig) UpdateChirpRedStatus(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetApiKey(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "authentication process failed")
		return
	}
	if !strings.EqualFold(apiKey, cfg.PolkaKey) {
		utils.RespondWithError(w, http.StatusUnauthorized, "invalid api key")
		return
	}

	var req webHookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error has occurred decoding request body. ERR: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if req.Event != "user.upgraded" {
		utils.RespondWithError(w, http.StatusNotFound, "not a valid event")
		return
	}
	userId, err := uuid.Parse(req.Data.UserId)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "not a valid user id")
		return
	}
	err = cfg.DbQueries.UpdateChirpyRed(r.Context(), userId)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "user doesn't exist")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
