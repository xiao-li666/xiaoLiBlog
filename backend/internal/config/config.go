package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	MySQLDSN       string
	JWTSecret      string
	RedisAddr      string
	FrontendOrigin string
	UploadDir      string
	Port           string
}

func Load(paths ...string) Config {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
		}
	}

	return Config{
		MySQLDSN:       env("MYSQL_DSN", ""),
		JWTSecret:      env("JWT_SECRET", "dev-secret-change-me"),
		RedisAddr:      env("REDIS_ADDR", ""),
		FrontendOrigin: env("FRONTEND_ORIGIN", "http://localhost:5173"),
		UploadDir:      env("UPLOAD_DIR", "uploads"),
		Port:           env("PORT", "8080"),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
