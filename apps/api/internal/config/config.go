package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	Region           string
	Bucket           string
	QueueURL         string
	JWTSecret        []byte
	AllowedOrigins   []string
	DatabaseURL      string
	RedisURL         string
	S3ForcePathStyle bool
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		Port:             getEnv("PORT", "8080"),
		Region:           os.Getenv("AWS_REGION"),
		Bucket:           os.Getenv("S3_BUCKET"),
		QueueURL:         os.Getenv("SQS_QUEUE_URL"),
		JWTSecret:        []byte(os.Getenv("JWT_SECRET")),
		AllowedOrigins:   parseOriginList(os.Getenv("ALLOWED_ORIGINS")),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		RedisURL:         os.Getenv("REDIS_URL"),
		S3ForcePathStyle: os.Getenv("S3_FORCE_PATH_STYLE") == "true",
	}

	if cfg.Region == "" {
		return Config{}, fmt.Errorf("AWS_REGION is required")
	}
	if cfg.Bucket == "" {
		return Config{}, fmt.Errorf("S3_BUCKET is required")
	}
	if cfg.QueueURL == "" {
		return Config{}, fmt.Errorf("SQS_QUEUE_URL is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET is required and must be at least 32 characters")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return Config{}, fmt.Errorf("ALLOWED_ORIGINS is required")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required")
	}

	return cfg, nil
}

func parseOriginList(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
