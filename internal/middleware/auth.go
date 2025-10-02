package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/spksupakorn/Currency-Converter/config"
	"github.com/spksupakorn/Currency-Converter/internal/repositories"
	"github.com/spksupakorn/Currency-Converter/internal/services"
	"github.com/spksupakorn/Currency-Converter/pkg/response"
)

func AuthRequired(cfg config.Config, userRepo repositories.UserRepository) gin.HandlerFunc {
	authSvc := services.NewAuthService(cfg, userRepo, nil)
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			response.Unauthorized(c, "unauthorized", "missing token")
			c.Abort()
			return
		}
		tokenStr := h
		// Support "Bearer <token>"
		if len(h) > 7 && h[:7] == "Bearer " {
			tokenStr = h[7:]
		}
		_, claims, err := authSvc.ParseToken(tokenStr)
		if err != nil {
			response.Unauthorized(c, "unauthorized", "invalid token")
			c.Abort()
			return
		}
		// Check token version against DB
		u, err := userRepo.FindByID(claims.UserID)
		if err != nil || u == nil {
			response.Unauthorized(c, "unauthorized", "user not found")
			c.Abort()
			return
		}
		if u.TokenVersion != claims.TokenVersion {
			response.Unauthorized(c, "unauthorized", "token revoked")
			c.Abort()
			return
		}

		c.Set("user_id", u.ID)
		c.Set("user_email", u.Email)
		c.Next()
	}
}
