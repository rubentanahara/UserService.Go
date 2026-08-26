package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestMiddleware(t *testing.T) {
	issuer := NewTokenIssuer("test-secret")
	validToken, err := issuer.Issue("user-1")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantUserID string
	}{
		{name: "missing header", authHeader: "", wantStatus: http.StatusUnauthorized},
		{name: "malformed header", authHeader: "Token abc", wantStatus: http.StatusUnauthorized},
		{name: "empty bearer", authHeader: "Bearer ", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", authHeader: "Bearer not-a-jwt", wantStatus: http.StatusUnauthorized},
		{name: "valid token", authHeader: "Bearer " + validToken, wantStatus: http.StatusOK, wantUserID: "user-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotUserID any
			r := gin.New()
			r.GET("/protected", Middleware(issuer), func(c *gin.Context) {
				gotUserID, _ = c.Get(userIDKey)
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantUserID != "" && gotUserID != tt.wantUserID {
				t.Fatalf("userID = %v, want %q", gotUserID, tt.wantUserID)
			}
		})
	}
}
