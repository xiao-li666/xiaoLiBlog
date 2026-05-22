package app

import (
	"log"
	"os"
	"time"

	"blogapp/backend/internal/config"
	"blogapp/backend/internal/database"
	"blogapp/backend/internal/handler"
	"blogapp/backend/internal/router"

	"github.com/redis/go-redis/v9"
)

func Run() {
	cfg := config.Load(".env", "../.env")
	if cfg.MySQLDSN == "" {
		log.Fatal("MYSQL_DSN is required")
	}
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("create upload dir: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := database.SeedDefaults(db); err != nil {
		log.Fatalf("seed: %v", err)
	}

	var rdb *redis.Client
	if cfg.RedisAddr != "" {
		rdb = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	}

	summaryAI := handler.NewSummaryGenerator(handler.SummaryAIConfig{
		APIKey:  cfg.DeepSeekAPIKey,
		BaseURL: cfg.DeepSeekBaseURL,
		Model:   cfg.DeepSeekModel,
		Timeout: time.Duration(cfg.DeepSeekTimeout) * time.Second,
	})
	h := handler.New(db, rdb, []byte(cfg.JWTSecret), cfg.UploadDir, handler.NewVerificationMailer(cfg), summaryAI)
	engine := router.New(cfg, h)

	log.Printf("backend listening on :%s", cfg.Port)
	if err := engine.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
