package auth

import (
	"errors"
	"net/http"
	"strings"
)

// Write a Unit Test for this!!!
func GetBearerToken(headers http.Header) (string, error) {
	val, ok := headers["Authorization"]
	if !ok || len(val) == 0 {
		return "", errors.New("no authorization header")
	}

	parts := strings.SplitN(val[0], " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("invalid authorization header")
	}

	return parts[1], nil
}
