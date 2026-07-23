package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	Region           string
	Bucket           string
	QueueURL         string
	JWTSecret        []byte
	AllowedOrigin    string
	DatabaseURL      string
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
		AllowedOrigin:    os.Getenv("ALLOWED_ORIGIN"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
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
	if cfg.AllowedOrigin == "" {
		return Config{}, fmt.Errorf("ALLOWED_ORIGIN is required")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
