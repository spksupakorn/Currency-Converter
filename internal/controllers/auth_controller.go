package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/spksupakorn/Currency-Converter/internal/services"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"
	"github.com/spksupakorn/Currency-Converter/pkg/response"
)

type AuthController struct {
	auth services.AuthService
	log  *logger.Logger
}

func NewAuthController(auth services.AuthService, log *logger.Logger) *AuthController {
	return &AuthController{auth: auth, log: log}
}

type registerReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthController) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid_request", err)
		return
	}
	if err := h.auth.Register(req.Email, req.Password); err != nil {
		response.BadRequest(c, "registration_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "registered"})
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthController) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid_request", err)
		return
	}
	token, _, err := h.auth.Login(req.Email, req.Password)
	if err != nil {
		response.Unauthorized(c, "login_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "bearer",
	})
}

func (h *AuthController) Logout(c *gin.Context) {
	userIDv, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "unauthorized", "missing user")
		return
	}
	userID := userIDv.(uint)
	if err := h.auth.Logout(userID); err != nil {
		response.InternalError(c, "logout_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
