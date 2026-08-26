package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestBodySizeLimit(t *testing.T) {
	tests := []struct {
		name    string
		limit   int64
		body    string
		wantErr bool
	}{
		{name: "within limit", limit: 10, body: "12345"},
		{name: "over limit", limit: 5, body: "123456789", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var readErr error
			r := gin.New()
			r.Use(bodySizeLimit(tt.limit))
			r.POST("/", func(c *gin.Context) {
				_, readErr = io.ReadAll(c.Request.Body)
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if (readErr != nil) != tt.wantErr {
				t.Fatalf("readErr = %v, wantErr = %v", readErr, tt.wantErr)
			}
		})
	}
}

func TestTimeout(t *testing.T) {
	const d = 50 * time.Millisecond

	var hasDeadline bool
	var deadline time.Time
	r := gin.New()
	r.Use(timeout(d))
	r.GET("/", func(c *gin.Context) {
		deadline, hasDeadline = c.Request.Context().Deadline()
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if !hasDeadline {
		t.Fatal("request context has no deadline")
	}
	if remaining := time.Until(deadline); remaining <= 0 || remaining > d {
		t.Fatalf("deadline %v from now, want between 0 and %v", remaining, d)
	}
}
