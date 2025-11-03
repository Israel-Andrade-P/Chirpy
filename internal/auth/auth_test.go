package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	plain_pwd := "cheese"
	hashed, err := HashPassword(plain_pwd)
	if err != nil {
		t.Fatalf("Error occurred during Testing: %v", err)
	}
	if hashed == plain_pwd {
		t.Errorf("Password not hashed! Plain text returned: %s", hashed)
	}

	//CheckPasswordHash func
	match, err := CheckPasswordHash(plain_pwd, hashed)
	if err != nil {
		t.Fatalf("Error occurred during Testing: %v", err)
	}
	if !match {
		t.Errorf("Plain password and hash didn't match!")
	}
	//Shouldn't match
	match, _ = CheckPasswordHash("wrongpwd", hashed)
	if match {
		t.Errorf("Passwords shouldn't match but they did!")
	}
}

func TestValidateJWT(t *testing.T) {
	expectedId := uuid.New()
	tokenSecret := "mysecret"
	expires := time.Second * 10
	jwtToken, err := MakeJWT(expectedId, tokenSecret, expires)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}
	resultId, err := ValidateJWT(jwtToken, tokenSecret)
	if err != nil {
		t.Fatalf("error occurred during testing: %v", err)
	}
	if resultId != expectedId {
		t.Errorf("token doesn't belong to logged in user")
	}

	//testing token expiration
	expiredJwt, err := MakeJWT(expectedId, tokenSecret, -1*time.Second)
	if err != nil {
		t.Fatalf("MakeJWT error: %v", err)
	}
	_, err = ValidateJWT(expiredJwt, tokenSecret)

	if err == nil {
		t.Errorf("expected token to be expired, but got no error")
	} else if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected expiration error, got this: %v", err)
	}
}
