package main

import (
	"github.com/spksupakorn/Currency-Converter/config"
	"github.com/spksupakorn/Currency-Converter/database"
	"github.com/spksupakorn/Currency-Converter/internal/server"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"
)

// @title           Currency Converter API
// @version         1.0
// @description     server for a currency converter application.
// @termsOfService  http://swagger.io/terms/
// @contact.name    API Support
// @contact.url     http://swagger.io/contact/
// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer {token}" to authenticate.
func main() {
	// Load config and logger
	cfg := config.Load()
	log := logger.New(cfg.Env)

	db := database.NewPostgresDatabase(cfg, log)
	err := db.MigrateDB()
	if err != nil {
		log.Fatal("Failed to migrate database", logger.Fields{"error": err.Error()})
	}

	server := server.NewGinServer(db, log, cfg)
	server.Start()
}
