package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/spksupakorn/Currency-Converter/config"
	"github.com/spksupakorn/Currency-Converter/database"
	"github.com/spksupakorn/Currency-Converter/docs"
	"github.com/spksupakorn/Currency-Converter/internal/controllers"
	"github.com/spksupakorn/Currency-Converter/internal/middleware"
	"github.com/spksupakorn/Currency-Converter/internal/router"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"
)

type ginServer struct {
	app *gin.Engine
	log *logger.Logger
	db  database.Database
	cfg config.Config
}

var (
	once sync.Once
	app  *ginServer
)

func NewGinServer(db database.Database, log *logger.Logger, cfg config.Config) Server {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	ginApp := gin.Default()

	once.Do(func() {
		app = &ginServer{
			app: ginApp,
			log: log,
			db:  db,
			cfg: cfg,
		}
	})
	return app
}

func (s *ginServer) Start() {
	s.app.Use(middleware.Recovery(s.log))
	s.app.Use(middleware.RequestID())
	s.app.Use(middleware.RequestLogger(s.log))
	s.app.Use(middleware.SecurityHeaders())
	s.app.Use(middleware.RateLimit(s.cfg))

	s.initRoutes()
	s.httpListenAndServe()
}

func (s *ginServer) httpListenAndServe() {
	// Start server in a goroutine
	port := fmt.Sprintf(":%d", s.cfg.Port)

	// HTTP server
	srv := &http.Server{
		Addr:              port,
		Handler:           s.app,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Graceful shutdown
	go func() {
		s.log.Info("server starting", logger.Fields{"port": s.cfg.Port})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Fatal("server error", zapErr(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	s.log.Info("server shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		s.log.Error("server shutdown error", zapErr(err))
	}
	s.log.Info("server exited")
}

func (s *ginServer) initRoutes() {
	router.NewRouter(s.db.ConnectDB(), s.log, s.cfg, s.app)

	// Swagger setup
	docs.SwaggerInfo.Title = "Currency Converter API Documentation"
	docs.SwaggerInfo.Description = "Server for a currency converter."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = fmt.Sprintf("%d", s.cfg.Port)
	if s.cfg.Port == 0 || s.cfg.Port == 8080 {
		docs.SwaggerInfo.Host = "localhost:8080"
	}
	docs.SwaggerInfo.BasePath = "/api/v1"

	// s.app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	s.app.GET("/openapi.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(docs.SwaggerInfo.ReadDoc()))
	})

	// Serve RapiDoc HTML
	s.app.GET("/docs", controllers.RapiDoc)
}

func zapErr(err error) logger.Fields {
	return logger.Fields{"error": err.Error()}
}
