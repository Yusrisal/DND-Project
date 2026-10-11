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
	UploadDir   string
	CORSOrigins []string
}

func Load() Config {
	return Config{
		Env:         env("APP_ENV", "development"),
		HTTPAddr:    env("HTTP_ADDR", ":8080"),
		DBDSN:       loadDSN(),
		WebDir:      os.Getenv("WEB_DIR"),
		UploadDir:   env("UPLOAD_DIR", "uploads"),
		CORSOrigins: strings.Split(env("CORS_ORIGINS", "http://localhost:5173"), ","),
	}
}

func loadDSN() string {
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		return dsn
	}

	host := env("DB_HOST", "")
	if host == "" {
		return "dnd:dnd@tcp(127.0.0.1:3306)/dnd?parseTime=true&charset=utf8mb4&loc=Local"
	}

	port := env("DB_PORT", "3306")
	user := env("DB_USER", "root")
	pass := os.Getenv("DB_PASSWORD")
	name := env("DB_NAME", "dnd")
	charset := env("DB_CHARSET", "utf8mb4")
	parseTime := env("DB_PARSE_TIME", "true")
	loc := env("DB_LOC", "Local")

	// user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=true&loc=Local
	var auth string
	if pass != "" {
		auth = user + ":" + pass
	} else {
		auth = user
	}

	return auth + "@tcp(" + host + ":" + port + ")/" + name +
		"?charset=" + charset + "&parseTime=" + parseTime + "&loc=" + loc
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
