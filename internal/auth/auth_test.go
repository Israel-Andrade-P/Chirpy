package auth

import (
	"errors"
	"net/http"
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

	match, err := CheckPasswordHash(plain_pwd, hashed)
	if err != nil {
		t.Fatalf("Error occurred during Testing: %v", err)
	}
	if !match {
		t.Errorf("Plain password and hash didn't match!")
	}
	//shouldn't match
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

	//wrong secret passed in
	_, err = ValidateJWT(jwtToken, "wrongsecret")
	if err == nil {
		t.Errorf("token validated even with wrong secret")
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

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantError error
	}{
		{
			name: "valid bearer token",
			headers: http.Header{
				"Authorization": []string{"Bearer token123"},
			},
			wantToken: "token123",
			wantError: nil,
		},
		{
			name:      "missing authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantError: errors.New("no authorization header"),
		},
		{
			name: "empty authorization header",
			headers: http.Header{
				"Authorization": []string{},
			},
			wantToken: "",
			wantError: errors.New("no authorization header"),
		},
		{
			name: "not bearer",
			headers: http.Header{
				"Authorization": []string{"Basic token123"},
			},
			wantToken: "",
			wantError: errors.New("invalid authorization header"),
		},
		{
			name: "missing token part",
			headers: http.Header{
				"Authorization": []string{"Bearer"},
			},
			wantToken: "",
			wantError: errors.New("invalid authorization header"),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			token, err := GetBearerToken(c.headers)

			if c.wantError == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.wantError != nil {
				if err == nil {
					t.Fatalf("expected error %v but got nil", err)
				}
				if c.wantError.Error() != err.Error() {
					t.Fatalf("expected error: %s got: %s", c.wantError.Error(), err.Error())
				}
			}
			if c.wantToken != token {
				t.Fatalf("expected token: %s got: %s", c.wantToken, token)
			}
		})
	}
}
