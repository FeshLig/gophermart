package handler

import (
	"errors"
	"net/http"

	"github.com/FeshLig/gophermart/internal/dto"
	"github.com/FeshLig/gophermart/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var creds dto.UserCredentials

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	token, err := h.authService.Register(c.Request.Context(), creds.Login, creds.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			c.Status(http.StatusConflict)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.SetCookie("token", token, 3600*24, "/", "", false, true)
	c.Status(http.StatusOK)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var creds dto.UserCredentials

	if err := c.ShouldBindJSON(&creds); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	token, err := h.authService.Login(c.Request.Context(), creds.Login, creds.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.Status(http.StatusUnauthorized)
			return
		}
		c.Status(http.StatusInternalServerError)
		return
	}

	c.SetCookie("token", token, 3600*24, "/", "", false, true)
	c.Status(http.StatusOK)
}
