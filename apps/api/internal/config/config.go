package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port     string
	Region   string
	Bucket   string
	QueueURL string
}

func Load() (Config, error) {
	cfg := Config{
		Port:     getEnv("PORT", "8080"),
		Region:   os.Getenv("AWS_REGION"),
		Bucket:   os.Getenv("S3_BUCKET"),
		QueueURL: os.Getenv("SQS_QUEUE_URL"),
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

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
