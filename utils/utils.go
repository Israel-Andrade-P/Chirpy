package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	ErrMsg string `json:"error"`
}

func RespondWithJson(w http.ResponseWriter, status int, payload any) {
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

func RespondWithError(w http.ResponseWriter, status int, errMsg string) {
	errPayload := ErrorResponse{ErrMsg: errMsg}
	RespondWithJson(w, status, errPayload)
}

func CleanUpMessage(message string) string {
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
