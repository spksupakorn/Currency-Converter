package database

import (
	"fmt"
	"time"

	"github.com/spksupakorn/Currency-Converter/config"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func Connect(cfg config.Config, log *logger.Logger) (*gorm.DB, error) {
	level := gormlogger.Warn
	newLogger := gormlogger.New(
		logger.NewGormZapWriter(log.Logger.Sugar(), level),
		gormlogger.Config{
			SlowThreshold: time.Second,
			LogLevel:      gormlogger.Warn,
			Colorful:      false,
		},
	)
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMODE, cfg.DBTimeZone,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
}
