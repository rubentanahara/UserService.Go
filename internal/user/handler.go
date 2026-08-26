package user

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rubentanahara/user_service/internal/auth"
)

const defaultListLimit = 20

type Handler struct {
	service Service
	logger  *slog.Logger
	tokens  *auth.TokenIssuer
}

func NewHandler(service Service, logger *slog.Logger, tokens *auth.TokenIssuer) *Handler {
	return &Handler{service: service, logger: logger, tokens: tokens}
}

func (h *Handler) Register(r gin.IRouter, loginRateLimit gin.HandlerFunc) {
	r.POST("/users", loginRateLimit, h.create)
	r.POST("/auth/login", loginRateLimit, h.login)

	protected := r.Group("/users", auth.Middleware(h.tokens))
	protected.GET("", h.list)
	protected.GET("/:id", h.getByID)
	protected.PUT("/:id", h.update)
	protected.PUT("/:id/password", h.changePassword)
	protected.DELETE("/:id", h.delete)
}

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *Handler) create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "bad_request"})
		return
	}

	u, err := h.service.Create(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (h *Handler) getByID(c *gin.Context) {
	u, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

type updateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

func (h *Handler) update(c *gin.Context) {
	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "bad_request"})
		return
	}

	u, err := h.service.Update(c.Request.Context(), c.Param("id"), req.Username, req.Email)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.Query("limit"), 10, 64)
	if limit <= 0 {
		limit = defaultListLimit
	}
	offset, _ := strconv.ParseInt(c.Query("offset"), 10, 64)

	users, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "bad_request"})
		return
	}

	u, err := h.service.Authenticate(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.writeError(c, err)
		return
	}

	token, err := h.tokens.Issue(u.ID.Hex())
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "issue token failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "code": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *Handler) changePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "bad_request"})
		return
	}

	err := h.service.ChangePassword(c.Request.Context(), c.Param("id"), req.OldPassword, req.NewPassword)
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "code": "not_found"})
	case errors.Is(err, ErrDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "code": "duplicate"})
	case errors.Is(err, ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "code": "invalid_credentials"})
	default:
		h.logger.ErrorContext(c.Request.Context(), "unhandled error", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "code": "internal_error"})
	}
}
