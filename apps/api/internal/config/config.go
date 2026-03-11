package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env  string

	DatabaseURL string

	GithubClientID     string
	GithubClientSecret string
	GithubRedirectURL  string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	JWTSecret      string
	JWTExpiryHours int

	R2AccountID       string
	R2Bucket          string
	R2AccessKeyID     string
	R2SecretAccessKey string

	CLIOAuthPort string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("config: .env not loaded (%v), using process environment", err)
	} else {
		log.Printf("config: loaded .env file")
	}

	cfg := &Config{
		Port:               getOrDefault("PORT", "8080"),
		Env:                getOrDefault("APP_ENV", "development"),
		DatabaseURL:        strings.TrimSpace(os.Getenv("DATABASE_URL")),
		GithubClientID:     strings.TrimSpace(os.Getenv("GITHUB_CLIENT_ID")),
		GithubClientSecret: strings.TrimSpace(os.Getenv("GITHUB_CLIENT_SECRET")),
		GithubRedirectURL:  getOrDefault("GITHUB_REDIRECT_URL", "http://localhost:8080/v1/auth/github/callback"),
		GoogleClientID:     strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),
		GoogleClientSecret: strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")),
		GoogleRedirectURL:  getOrDefault("GOOGLE_REDIRECT_URL", "http://localhost:8080/v1/auth/google/callback"),
		JWTSecret:          strings.TrimSpace(os.Getenv("JWT_SECRET")),
		R2AccountID:        strings.TrimSpace(os.Getenv("R2_ACCOUNT_ID")),
		R2Bucket:           strings.TrimSpace(os.Getenv("R2_BUCKET")),
		R2AccessKeyID:      strings.TrimSpace(os.Getenv("R2_ACCESS_KEY_ID")),
		R2SecretAccessKey:  strings.TrimSpace(os.Getenv("R2_SECRET_ACCESS_KEY")),
		CLIOAuthPort:       getOrDefault("CLI_OAUTH_PORT", "9876"),
	}

	jwtHours, err := intFromEnv("JWT_EXPIRY_HOURS", 720)
	if err != nil {
		return nil, err
	}
	cfg.JWTExpiryHours = jwtHours

	missing := make([]string, 0)
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if cfg.GithubClientID == "" || cfg.GithubClientSecret == "" {
		log.Printf("config: github oauth not fully configured")
	}
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
		log.Printf("config: google oauth not fully configured")
	}
	if cfg.R2AccountID == "" || cfg.R2Bucket == "" || cfg.R2AccessKeyID == "" || cfg.R2SecretAccessKey == "" {
		log.Printf("config: r2 storage not fully configured")
	}

	return cfg, nil
}

func getOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func intFromEnv(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid integer for %s: %w", key, err)
	}
	return value, nil
}
