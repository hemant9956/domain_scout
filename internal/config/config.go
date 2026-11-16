package config

import "os"

type Config struct {
	HTTPAddress       string
	SSHTimeoutSeconds int
	MaxRequestBytes   int64
}

func Load() Config {
	return Config{
		HTTPAddress:       env("HTTP_ADDRESS", ":8080"),
		SSHTimeoutSeconds: 15,
		MaxRequestBytes:   1 << 20,
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
