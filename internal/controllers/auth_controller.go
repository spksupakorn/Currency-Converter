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
	Username string `json:"username"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        registerReq  body      registerReq  true  "Register Request"
// @Success      201          {object}  map[string]string
// @Failure      400          {object}  response.ErrorResponse
// @Failure      500          {object}  response.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthController) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "invalid_request", err)
		return
	}
	if err := h.auth.Register(req.Username, req.Email, req.Password); err != nil {
		response.BadRequest(c, "registration_failed", err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "registered"})
}

type loginReq struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login godoc
// @Summary      Login a user
// @Description  Login a user with email and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        loginReq  body      loginReq  true  "Login Request"
// @Success      200       {object}  map[string]string
// @Failure      400       {object}  response.ErrorResponse
// @Failure	  401       {object}  response.ErrorResponse
// @Failure	  500       {object}  response.ErrorResponse
// @Router       /auth/login [post]
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

// Logout godoc
// @Summary      Logout a user
// @Description  Logout a user by invalidating their token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  response.ErrorResponse
// @Failure      500  {object}  response.ErrorResponse
// @Router       /auth/logout [post]
// @Security     BearerAuth
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
