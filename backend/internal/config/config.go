package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
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
	return Config{HTTPAddr: env("HTTP_ADDR", ":8080"), DatabaseURL: u.String()}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
