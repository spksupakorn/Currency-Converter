package database

import (
	"fmt"
	"sync"

	"github.com/spksupakorn/Currency-Converter/config"
	"github.com/spksupakorn/Currency-Converter/internal/models"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type postgresDatabase struct {
	Db *gorm.DB
}

var (
	once       sync.Once
	dbInstance *postgresDatabase
)

func NewPostgresDatabase(cfg config.Config, log *logger.Logger) Database {
	once.Do(func() {
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMODE, cfg.DBTimeZone,
		)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		})
		if err != nil {
			panic("failed to connect database")
		}

		fmt.Println("💰 Successfully connected to the database")

		dbInstance = &postgresDatabase{Db: db}
	})

	return dbInstance
}

func (p *postgresDatabase) ConnectDB() *gorm.DB {
	return p.Db
}

func (p *postgresDatabase) MigrateDB() error {
	return p.Db.AutoMigrate(
		&models.User{},
		&models.Rate{},
	)
}
