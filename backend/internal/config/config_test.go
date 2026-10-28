package config

import (
	"net/url"
	"testing"
)

func TestLoadRequiresPassword(t *testing.T) {
	t.Setenv("POSTGRES_PASSWORD", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error for missing password")
	}
}

func TestLoadEscapesCredentials(t *testing.T) {
	t.Setenv("POSTGRES_PASSWORD", "p@ss:/?# word")
	t.Setenv("POSTGRES_USER", "test@user")
	t.Setenv("POSTGRES_HOST", "::1")
	t.Setenv("POSTGRES_PORT", "55432")
	t.Setenv("POSTGRES_DB", "expense_test")
	t.Setenv("HTTP_ADDR", ":18080")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(cfg.DatabaseURL)
	if err != nil {
		t.Fatal(err)
	}
	password, _ := u.User.Password()
	if password != "p@ss:/?# word" || u.User.Username() != "test@user" || u.Host != "[::1]:55432" || u.Path != "/expense_test" || u.Query().Get("timezone") != "UTC" || cfg.HTTPAddr != ":18080" {
		t.Fatal("config did not preserve environment values")
	}
}

func TestAuthConfig(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if _, err := LoadAuth(); err == nil {
		t.Fatal("missing secret accepted")
	}
	t.Setenv("JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("JWT_ACCESS_TTL", "15m")
	t.Setenv("REFRESH_TOKEN_TTL", "720h")
	if _, err := LoadAuth(); err != nil {
		t.Fatal(err)
	}
	for _, ttl := range []string{"invalid", "0s", "500ms", "2h"} {
		t.Setenv("JWT_ACCESS_TTL", ttl)
		if _, err := LoadAuth(); err == nil {
			t.Fatalf("invalid TTL %s accepted", ttl)
		}
	}
	t.Setenv("JWT_ACCESS_TTL", "15m")
	t.Setenv("REFRESH_TOKEN_TTL", "1m")
	if _, err := LoadAuth(); err == nil {
		t.Fatal("short refresh TTL accepted")
	}
}
