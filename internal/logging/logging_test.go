package logging

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)
	logger.Info("hello")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("decode log line: %v, body = %s", err, buf.String())
	}
	if entry["msg"] != "hello" {
		t.Fatalf("msg = %v, want %q", entry["msg"], "hello")
	}
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		incomingReqID string
	}{
		{name: "generates request id when missing"},
		{name: "echoes incoming request id", incomingReqID: "given-id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf)

			r := gin.New()
			r.Use(Middleware(logger))
			r.GET("/ok", func(c *gin.Context) {
				c.Status(http.StatusTeapot)
			})

			req := httptest.NewRequest(http.MethodGet, "/ok", nil)
			if tt.incomingReqID != "" {
				req.Header.Set(requestIDHeader, tt.incomingReqID)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			respID := rec.Header().Get(requestIDHeader)
			if respID == "" {
				t.Fatal("response missing request id header")
			}
			if tt.incomingReqID != "" && respID != tt.incomingReqID {
				t.Fatalf("request id = %q, want echoed %q", respID, tt.incomingReqID)
			}

			var entry map[string]any
			if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
				t.Fatalf("decode log line: %v, body = %s", err, buf.String())
			}
			if entry["request_id"] != respID {
				t.Fatalf("logged request_id = %v, want %q", entry["request_id"], respID)
			}

			httpAttrs, ok := entry["http"].(map[string]any)
			if !ok {
				t.Fatalf("missing http group in log entry: %v", entry)
			}
			if got := httpAttrs["status"]; got != float64(http.StatusTeapot) {
				t.Fatalf("logged status = %v, want %d", got, http.StatusTeapot)
			}
			if got := httpAttrs["path"]; got != "/ok" {
				t.Fatalf("logged path = %v, want %q", got, "/ok")
			}
		})
	}
}

func TestNewRequestID(t *testing.T) {
	id := newRequestID()
	if len(id) != 32 {
		t.Fatalf("len(id) = %d, want 32 (16 hex-encoded bytes)", len(id))
	}
	if strings.ContainsAny(id, "GHIJKLMNOPQRSTUVWXYZ") {
		t.Fatalf("id %q is not lowercase hex", id)
	}
}
