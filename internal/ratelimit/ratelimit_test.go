package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter(rps float64, burst int) *gin.Engine {
	r := gin.New()
	r.Use(PerIP(rps, burst))
	r.GET("/ok", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func request(r http.Handler, remoteAddr string) int {
	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	req.RemoteAddr = remoteAddr
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec.Code
}

func TestPerIP_AllowsWithinBurst(t *testing.T) {
	r := newTestRouter(1, 3)

	for i := range 3 {
		if code := request(r, "1.2.3.4:1111"); code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want %d", i, code, http.StatusOK)
		}
	}
}

func TestPerIP_RejectsOverBurst(t *testing.T) {
	r := newTestRouter(1, 3)

	for range 3 {
		request(r, "1.2.3.4:1111")
	}
	if code := request(r, "1.2.3.4:1111"); code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", code, http.StatusTooManyRequests)
	}
}

func TestPerIP_TracksIndependentlyPerIP(t *testing.T) {
	r := newTestRouter(1, 1)

	if code := request(r, "1.2.3.4:1111"); code != http.StatusOK {
		t.Fatalf("ip1 request: status = %d, want %d", code, http.StatusOK)
	}
	if code := request(r, "5.6.7.8:2222"); code != http.StatusOK {
		t.Fatalf("ip2 request: status = %d, want %d", code, http.StatusOK)
	}
	if code := request(r, "1.2.3.4:1111"); code != http.StatusTooManyRequests {
		t.Fatalf("ip1 second request: status = %d, want %d", code, http.StatusTooManyRequests)
	}
}
