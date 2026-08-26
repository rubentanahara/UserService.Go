package user

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
	logger  *slog.Logger
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Register(r gin.IRouter) {
	r.POST("/users", h.create)
	r.GET("/users/:id", h.getByID)
	r.PUT("/users/:id", h.update)
	r.DELETE("/users/:id", h.delete)
}

type createUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *Handler) create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.service.Create(c.Request.Context(), req.Username, req.Email, req.Password)
	if errors.Is(err, ErrDuplicate) {
		c.JSON(http.StatusConflict, gin.H{"error": ErrDuplicate.Error()})
		return
	}
	if err != nil {
		h.logger.Error("create user failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}
	c.JSON(http.StatusCreated, u)
}

func (h *Handler) getByID(c *gin.Context) {
	u, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		h.logger.Error("get user failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not get user"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.service.Update(c.Request.Context(), c.Param("id"), req.Username, req.Email)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if errors.Is(err, ErrDuplicate) {
		c.JSON(http.StatusConflict, gin.H{"error": ErrDuplicate.Error()})
		return
	}
	if err != nil {
		h.logger.Error("update user failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update user"})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) delete(c *gin.Context) {
	err := h.service.Delete(c.Request.Context(), c.Param("id"))
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if err != nil {
		h.logger.Error("delete user failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete user"})
		return
	}
	c.Status(http.StatusNoContent)
}
