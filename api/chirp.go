package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/Israel-Andrade-P/Chirpy.git/internal/auth"
	"github.com/Israel-Andrade-P/Chirpy.git/internal/database"
	"github.com/Israel-Andrade-P/Chirpy.git/utils"
	"github.com/google/uuid"
)

type (
	chirpRequest struct {
		Body string `json:"body"`
	}
	chirpResponse struct {
		ID        uuid.UUID `json:"id"`
		Body      string    `json:"body"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		UserID    uuid.UUID `json:"user_id"`
	}
)

func (cfg *Apiconfig) SaveChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Please autenticate yourself")
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Your token is invalid, get a new one")
		return
	}
	var chirpReq chirpRequest
	if err := json.NewDecoder(r.Body).Decode(&chirpReq); err != nil {
		log.Printf("Error has occurred decoding request body. ERR: %v\n", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Error")
		return
	}
	if len(chirpReq.Body) > 140 {
		utils.RespondWithError(w, http.StatusBadRequest, "Chirp is too long.")
		return
	}
	cleanMessage := utils.CleanUpMessage(chirpReq.Body)
	chirp, err := cfg.DbQueries.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleanMessage, UserID: userId})
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid Chirp")
		return
	}
	c := chirpResponse{ID: chirp.ID, Body: chirp.Body, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, UserID: chirp.UserID}
	utils.RespondWithJson(w, http.StatusCreated, c)
}

func (cfg *Apiconfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	authorId := r.URL.Query().Get("author_id")
	sortParam := r.URL.Query().Get("sort")

	chirpsRes := make([]chirpResponse, 0)
	if authorId != "" {
		userId, err := uuid.Parse(authorId)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "invalid id")
			return
		}
		chirps, err := cfg.DbQueries.GetChirpsByUser(r.Context(), userId)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "no chirps")
			return
		}
		for _, chirp := range chirps {
			chirpsRes = append(chirpsRes, chirpResponse(chirp))
		}
		if sortParam == "desc" {
			sort.Slice(chirpsRes, func(i, j int) bool {
				return chirpsRes[i].CreatedAt.After(chirpsRes[j].CreatedAt)
			})
		}
		utils.RespondWithJson(w, http.StatusOK, chirpsRes)
		return
	}
	chirps, err := cfg.DbQueries.GetAllChirps(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "No Chirps")
	}
	for _, chirp := range chirps {
		chirpsRes = append(chirpsRes, chirpResponse(chirp))
	}
	if sortParam == "desc" {
		sort.Slice(chirpsRes, func(i, j int) bool {
			return chirpsRes[i].CreatedAt.After(chirpsRes[j].CreatedAt)
		})
	}
	utils.RespondWithJson(w, http.StatusOK, chirpsRes)
}

func (cfg *Apiconfig) GetChirp(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("chirpID")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	chirp, err := cfg.DbQueries.GetChirp(r.Context(), parsedId)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "DB Error")
		return
	}
	chirpRes := chirpResponse{ID: chirp.ID, Body: chirp.Body, CreatedAt: chirp.CreatedAt, UpdatedAt: chirp.UpdatedAt, UserID: chirp.UserID}
	utils.RespondWithJson(w, http.StatusOK, chirpRes)
}

func (cfg *Apiconfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Please autenticate yourself")
		return
	}
	userId, err := auth.ValidateJWT(token, cfg.Secret)
	if err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "your token is invalid, get a new one")
		return
	}
	id := r.PathValue("chirpID")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "invalid ID")
		return
	}
	chirp, err := cfg.DbQueries.GetChirp(r.Context(), parsedId)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "chirp doesn't exist")
		return
	}
	if chirp.UserID != userId {
		utils.RespondWithError(w, http.StatusForbidden, "that chirp doesn't belong to you")
		return
	}
	cfg.DbQueries.DeleteChirp(r.Context(), chirp.ID)
	w.WriteHeader(http.StatusNoContent)
}
