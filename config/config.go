package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                 string
	Port                int
	DBHost              string
	DBPort              int
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMODE           string
	DBTimeZone          string
	JWTSecret           string
	JWTExpiry           time.Duration
	RateBaseCurrency    string
	RateRefreshInterval time.Duration
	HTTPClientTimeout   time.Duration
	RateLimitRequests   int
	RateLimitWindow     time.Duration
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		Env:                 getEnv("APP_ENV", "development"),
		Port:                getInt("PORT", 8080),
		DBHost:              getEnv("DB_HOST", "localhost"),
		DBPort:              getInt("DB_PORT", 5432),
		DBUser:              getEnv("DB_USER", "postgres"),
		DBPassword:          mustEnv("DB_PASSWORD"),
		DBName:              getEnv("DB_NAME", "currencydb"),
		DBSSLMODE:           getEnv("DB_SSLMODE", "disable"),
		DBTimeZone:          getEnv("DB_TIMEZONE", "UTC"),
		JWTSecret:           mustEnv("JWT_SECRET"),
		JWTExpiry:           getDuration("JWT_EXPIRY", 24*time.Hour),
		RateBaseCurrency:    strings.ToUpper(getEnv("RATE_BASE_CURRENCY", "USD")),
		RateRefreshInterval: getDuration("RATE_REFRESH_INTERVAL", 6*time.Hour),
		HTTPClientTimeout:   getDuration("HTTP_CLIENT_TIMEOUT", 10*time.Second),
		RateLimitRequests:   getInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:     getDuration("RATE_LIMIT_WINDOW", time.Minute),
	}
	return cfg
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing required env: " + key)
	}
	return v
}

func getDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func getInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
