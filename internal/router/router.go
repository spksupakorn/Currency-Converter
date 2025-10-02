package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spksupakorn/Currency-Converter/config"
	"github.com/spksupakorn/Currency-Converter/internal/controllers"
	"github.com/spksupakorn/Currency-Converter/internal/middleware"
	"github.com/spksupakorn/Currency-Converter/internal/repositories"
	"github.com/spksupakorn/Currency-Converter/internal/services"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, log *logger.Logger, cfg config.Config, route *gin.Engine) {
	// Health
	route.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Repos
	userRepo := repositories.NewUserRepository(db)
	rateRepo := repositories.NewRateRepository(db)

	// Services
	authSvc := services.NewAuthService(cfg, userRepo, log)
	rateSvc := services.NewRateService(cfg, rateRepo, log)
	// Start rates background refresher
	rateSvc.StartBackgroundRefresh(context.Background())

	// API v1
	v1 := route.Group("/api/v1")
	{
		authH := controllers.NewAuthController(authSvc, log)
		v1.POST("/auth/register", authH.Register)
		v1.POST("/auth/login", authH.Login)
		v1.POST("/auth/logout", middleware.AuthRequired(cfg, userRepo), authH.Logout)

		rateH := controllers.NewRateController(rateSvc, log)
		protected := v1.Group("/")
		protected.Use(middleware.AuthRequired(cfg, userRepo))
		{
			protected.GET("/rates", rateH.GetRates)          // ?base=USD
			protected.GET("/convert", rateH.ConvertCurrency) // ?from=USD&to=THB&amount=123.45
		}
	}
}
