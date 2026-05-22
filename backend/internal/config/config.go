package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	MySQLDSN        string
	JWTSecret       string
	RedisAddr       string
	SMTPHost        string
	SMTPPort        string
	SMTPUsername    string
	SMTPPassword    string
	SMTPFrom        string
	FrontendOrigin  string
	UploadDir       string
	Port            string
	DeepSeekAPIKey  string
	DeepSeekBaseURL string
	DeepSeekModel   string
	DeepSeekTimeout int
}

func Load(paths ...string) Config {
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
		}
	}

	return Config{
		MySQLDSN:        env("MYSQL_DSN", ""),
		JWTSecret:       env("JWT_SECRET", "dev-secret-change-me"),
		RedisAddr:       env("REDIS_ADDR", ""),
		SMTPHost:        env("SMTP_HOST", ""),
		SMTPPort:        env("SMTP_PORT", ""),
		SMTPUsername:    env("SMTP_USERNAME", ""),
		SMTPPassword:    env("SMTP_PASSWORD", ""),
		SMTPFrom:        env("SMTP_FROM", ""),
		FrontendOrigin:  env("FRONTEND_ORIGIN", "http://localhost:5173"),
		UploadDir:       env("UPLOAD_DIR", "uploads"),
		Port:            env("PORT", "8080"),
		DeepSeekAPIKey:  env("DEEPSEEK_API_KEY", ""),
		DeepSeekBaseURL: env("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
		DeepSeekModel:   env("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		DeepSeekTimeout: envInt("DEEPSEEK_TIMEOUT_SECONDS", 30),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			return parsed
		}
	}
	return fallback
}
