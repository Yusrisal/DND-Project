package config

import (
	"os"
	"strings"
)

type Config struct {
	Env         string
	HTTPAddr    string
	DBDSN       string
	WebDir      string
	CORSOrigins []string
}

func Load() Config {
	return Config{
		Env:         env("APP_ENV", "development"),
		HTTPAddr:    env("HTTP_ADDR", ":8080"),
		DBDSN:       env("DB_DSN", "dnd:dnd@tcp(127.0.0.1:3306)/dnd?parseTime=true&charset=utf8mb4&loc=Local"),
		WebDir:      os.Getenv("WEB_DIR"),
		CORSOrigins: strings.Split(env("CORS_ORIGINS", "http://localhost:5173"), ","),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
