package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"
)

type Config struct {
	HTTPAddr          string
	DatabaseURL       string
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string
	R2PutURLTTL       time.Duration
	R2GetURLTTL       time.Duration
	R2Enabled         bool
}

func Load() (Config, error) {
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		return Config{}, fmt.Errorf("POSTGRES_PASSWORD is required")
	}
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(env("POSTGRES_USER", "expense_app"), password),
		Host:   net.JoinHostPort(env("POSTGRES_HOST", "localhost"), env("POSTGRES_PORT", "5432")),
		Path:   "/" + env("POSTGRES_DB", "expense_app"),
	}
	q := u.Query()
	q.Set("sslmode", env("POSTGRES_SSLMODE", "disable"))
	q.Set("timezone", "UTC")
	u.RawQuery = q.Encode()
	putTTL, err := time.ParseDuration(env("R2_PUT_URL_TTL", "10m"))
	if err != nil || putTTL <= 0 || putTTL > 1*time.Hour {
		return Config{}, fmt.Errorf("R2_PUT_URL_TTL must be a duration between 0 and 1h")
	}
	getTTL, err := time.ParseDuration(env("R2_GET_URL_TTL", "15m"))
	if err != nil || getTTL <= 0 || getTTL > 1*time.Hour {
		return Config{}, fmt.Errorf("R2_GET_URL_TTL must be a duration between 0 and 1h")
	}
	r2 := []string{os.Getenv("R2_ACCOUNT_ID"), os.Getenv("R2_ACCESS_KEY_ID"), os.Getenv("R2_SECRET_ACCESS_KEY"), os.Getenv("R2_BUCKET")}
	r2Count := 0
	for _, value := range r2 {
		if value != "" {
			r2Count++
		}
	}
	if r2Count != 0 && r2Count != len(r2) {
		return Config{}, fmt.Errorf("R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, and R2_BUCKET must be configured together")
	}
	return Config{
		HTTPAddr: env("HTTP_ADDR", ":8080"), DatabaseURL: u.String(),
		R2AccountID: r2[0], R2AccessKeyID: r2[1], R2SecretAccessKey: r2[2], R2Bucket: r2[3],
		R2PutURLTTL: putTTL, R2GetURLTTL: getTTL, R2Enabled: r2Count == len(r2),
	}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
