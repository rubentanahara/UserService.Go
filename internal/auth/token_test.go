package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenIssuer_IssueParse(t *testing.T) {
	issuer := NewTokenIssuer("test-secret")

	token, err := issuer.Issue("user-1")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	userID, err := issuer.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if userID != "user-1" {
		t.Fatalf("userID = %q, want %q", userID, "user-1")
	}
}

func TestTokenIssuer_Parse(t *testing.T) {
	issuer := NewTokenIssuer("test-secret")

	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "user-1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	wrongSecret, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: "user-1",
	}).SignedString([]byte("other-secret"))
	if err != nil {
		t.Fatalf("sign wrong-secret token: %v", err)
	}

	wrongAlg, err := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.RegisteredClaims{
		Subject: "user-1",
	}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign wrong-alg token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{name: "expired", token: expired, wantErr: ErrInvalidToken},
		{name: "wrong secret", token: wrongSecret, wantErr: ErrInvalidToken},
		{name: "wrong signing method", token: wrongAlg, wantErr: ErrInvalidToken},
		{name: "malformed", token: "not-a-jwt", wantErr: ErrInvalidToken},
		{name: "empty", token: "", wantErr: ErrInvalidToken},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := issuer.Parse(tt.token)
			if err != tt.wantErr {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
