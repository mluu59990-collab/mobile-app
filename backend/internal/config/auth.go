package config

import (
	"fmt"
	"os"
	"time"
)

type Auth struct {
	Secret            string
	Issuer            string
	Audience          string
	GoogleClientID    string
	FacebookAppID     string
	FacebookAppSecret string
	AccessTTL         time.Duration
	RefreshTTL        time.Duration
}

// Separate from Load so migration commands do not need signing credentials.
func LoadAuth() (Auth, error) {
	c := Auth{
		Secret:            os.Getenv("JWT_SECRET"),
		Issuer:            env("JWT_ISSUER", "expense-app"),
		Audience:          env("JWT_AUDIENCE", "expense-app-mobile"),
		GoogleClientID:    os.Getenv("GOOGLE_CLIENT_ID"),
		FacebookAppID:     os.Getenv("FACEBOOK_APP_ID"),
		FacebookAppSecret: os.Getenv("FACEBOOK_APP_SECRET"),
	}
	if len(c.Secret) < 32 {
		return Auth{}, fmt.Errorf("JWT_SECRET must contain at least 32 bytes")
	}
	var err error
	c.AccessTTL, err = time.ParseDuration(env("JWT_ACCESS_TTL", "15m"))
	if err != nil || c.AccessTTL < time.Second || c.AccessTTL > time.Hour {
		return Auth{}, fmt.Errorf("JWT_ACCESS_TTL must be between 1s and 1h")
	}
	c.RefreshTTL, err = time.ParseDuration(env("REFRESH_TOKEN_TTL", "720h"))
	if err != nil || c.RefreshTTL <= c.AccessTTL {
		return Auth{}, fmt.Errorf("REFRESH_TOKEN_TTL must exceed JWT_ACCESS_TTL")
	}
	return c, nil
}
