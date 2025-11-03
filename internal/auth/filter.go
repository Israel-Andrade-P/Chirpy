package auth

import (
	"errors"
	"net/http"
	"strings"
)

// Write a Unit Test for this!!!
func GetBearerToken(headers http.Header) (string, error) {
	token := ""
	if val, ok := headers["Authorization"]; ok {
		token = strings.Split(val[0], " ")[1]
		return token, nil
	}
	return "", errors.New("no authorization header")
}
