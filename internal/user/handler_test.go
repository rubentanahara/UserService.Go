package user

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"

	"github.com/rubentanahara/user_service/internal/auth"
)

const testTokenSecret = "test-secret"

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestHandler(t *testing.T) (*Handler, *MockService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	svc := NewMockService(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewHandler(svc, logger, auth.NewTokenIssuer(testTokenSecret)), svc
}

func performRequest(r http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockUser   *User
		mockErr    error
		wantStatus int
	}{
		{
			name:       "created",
			body:       `{"username":"jane","email":"jane@example.com","password":"password123"}`,
			mockUser:   &User{Username: "jane", Email: "jane@example.com"},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "duplicate",
			body:       `{"username":"jane","email":"jane@example.com","password":"password123"}`,
			mockErr:    ErrDuplicate,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "internal error",
			body:       `{"username":"jane","email":"jane@example.com","password":"password123"}`,
			mockErr:    errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "bad request",
			body:       `{"username":"jane"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, svc := newTestHandler(t)
			r := gin.New()
			r.POST("/users", h.create)

			if tt.wantStatus != http.StatusBadRequest {
				svc.EXPECT().Create(gomock.Any(), "jane", "jane@example.com", "password123").
					Return(tt.mockUser, tt.mockErr)
			}

			rec := performRequest(r, http.MethodPost, "/users", []byte(tt.body))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		mockUser   *User
		mockErr    error
		wantStatus int
	}{
		{name: "found", mockUser: &User{Username: "jane"}, wantStatus: http.StatusOK},
		{name: "not found", mockErr: ErrNotFound, wantStatus: http.StatusNotFound},
		{name: "internal error", mockErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, svc := newTestHandler(t)
			r := gin.New()
			r.GET("/users/:id", h.getByID)

			svc.EXPECT().GetByID(gomock.Any(), "1").Return(tt.mockUser, tt.mockErr)

			rec := performRequest(r, http.MethodGet, "/users/1", nil)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockUser   *User
		mockErr    error
		wantStatus int
	}{
		{
			name:       "updated",
			body:       `{"username":"jane","email":"jane@example.com"}`,
			mockUser:   &User{Username: "jane", Email: "jane@example.com"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found",
			body:       `{"username":"jane","email":"jane@example.com"}`,
			mockErr:    ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "duplicate",
			body:       `{"username":"jane","email":"jane@example.com"}`,
			mockErr:    ErrDuplicate,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "internal error",
			body:       `{"username":"jane","email":"jane@example.com"}`,
			mockErr:    errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "bad request",
			body:       `{"username":"jane"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, svc := newTestHandler(t)
			r := gin.New()
			r.PUT("/users/:id", h.update)

			if tt.wantStatus != http.StatusBadRequest {
				svc.EXPECT().Update(gomock.Any(), "1", "jane", "jane@example.com").
					Return(tt.mockUser, tt.mockErr)
			}

			rec := performRequest(r, http.MethodPut, "/users/1", []byte(tt.body))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		mockErr    error
		wantStatus int
	}{
		{name: "deleted", wantStatus: http.StatusNoContent},
		{name: "not found", mockErr: ErrNotFound, wantStatus: http.StatusNotFound},
		{name: "internal error", mockErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, svc := newTestHandler(t)
			r := gin.New()
			r.DELETE("/users/:id", h.delete)

			svc.EXPECT().Delete(gomock.Any(), "1").Return(tt.mockErr)

			rec := performRequest(r, http.MethodDelete, "/users/1", nil)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		mockErr    error
		wantStatus int
	}{
		{name: "listed", wantStatus: http.StatusOK},
		{name: "internal error", mockErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, svc := newTestHandler(t)
			r := gin.New()
			r.GET("/users", h.list)

			svc.EXPECT().List(gomock.Any(), int64(defaultListLimit), int64(0)).Return([]*User{}, tt.mockErr)

			rec := performRequest(r, http.MethodGet, "/users", nil)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockUser   *User
		mockErr    error
		wantStatus int
	}{
		{
			name:       "authenticated",
			body:       `{"email":"jane@example.com","password":"password123"}`,
			mockUser:   &User{Username: "jane", Email: "jane@example.com"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid credentials",
			body:       `{"email":"jane@example.com","password":"wrong"}`,
			mockErr:    ErrInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad request",
			body:       `{"email":"jane@example.com"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, svc := newTestHandler(t)
			r := gin.New()
			r.POST("/auth/login", h.login)

			if tt.wantStatus != http.StatusBadRequest {
				svc.EXPECT().Authenticate(gomock.Any(), "jane@example.com", gomock.Any()).
					Return(tt.mockUser, tt.mockErr)
			}

			rec := performRequest(r, http.MethodPost, "/auth/login", []byte(tt.body))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHandler_ChangePassword(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "changed",
			body:       `{"old_password":"password123","new_password":"newpassword123"}`,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid credentials",
			body:       `{"old_password":"wrong","new_password":"newpassword123"}`,
			mockErr:    ErrInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "not found",
			body:       `{"old_password":"password123","new_password":"newpassword123"}`,
			mockErr:    ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "bad request",
			body:       `{"old_password":"password123"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, svc := newTestHandler(t)
			r := gin.New()
			r.PUT("/users/:id/password", h.changePassword)

			if tt.wantStatus != http.StatusBadRequest {
				svc.EXPECT().ChangePassword(gomock.Any(), "1", gomock.Any(), gomock.Any()).Return(tt.mockErr)
			}

			rec := performRequest(r, http.MethodPut, "/users/1/password", []byte(tt.body))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestHandler_Register_Auth(t *testing.T) {
	h, svc := newTestHandler(t)
	r := gin.New()
	h.Register(r, func(c *gin.Context) { c.Next() })

	t.Run("public create needs no token", func(t *testing.T) {
		svc.EXPECT().Create(gomock.Any(), "jane", "jane@example.com", "password123").
			Return(&User{Username: "jane"}, nil)

		rec := performRequest(r, http.MethodPost, "/users",
			[]byte(`{"username":"jane","email":"jane@example.com","password":"password123"}`))
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
		}
	})

	t.Run("protected route rejects missing token", func(t *testing.T) {
		rec := performRequest(r, http.MethodGet, "/users/1", nil)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
	})

	t.Run("protected route accepts valid token", func(t *testing.T) {
		token, err := h.tokens.Issue("1")
		if err != nil {
			t.Fatalf("issue token: %v", err)
		}
		svc.EXPECT().GetByID(gomock.Any(), "1").Return(&User{Username: "jane"}, nil)

		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
		}
	})
}
