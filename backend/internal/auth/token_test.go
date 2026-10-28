package auth

import (
	"strings"
	"testing"
	"time"

	"expense-app/backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestAccessValidation(t *testing.T) {
	tokens := Tokens{Config: config.Auth{Secret: strings.Repeat("a", 32), Issuer: "test", Audience: "mobile", AccessTTL: time.Minute}}
	id := uuid.NewString()
	raw, err := tokens.Access(id)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := tokens.Parse(raw); err != nil || got != id {
		t.Fatalf("valid token rejected: %v", err)
	}
	now := time.Now()
	base := jwt.RegisteredClaims{Subject: id, Issuer: "test", Audience: jwt.ClaimStrings{"mobile"}, ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute))}
	for _, name := range []string{"expired", "wrong issuer", "wrong audience", "missing expiry", "future nbf", "invalid subject", "missing jti", "wrong algorithm", "wrong secret"} {
		t.Run(name, func(t *testing.T) {
			claims := base
			method := jwt.SigningMethodHS256
			secret := tokens.Config.Secret
			switch name {
			case "expired":
				claims.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Minute))
			case "wrong issuer":
				claims.Issuer = "other"
			case "wrong audience":
				claims.Audience = jwt.ClaimStrings{"other"}
			case "missing expiry":
				claims.ExpiresAt = nil
			case "future nbf":
				claims.NotBefore = jwt.NewNumericDate(now.Add(time.Hour))
			case "invalid subject":
				claims.Subject = "abc"
			case "missing jti":
				claims.ID = ""
			case "wrong algorithm":
				method = jwt.SigningMethodHS384
			case "wrong secret":
				secret = strings.Repeat("b", 32)
			}
			bad, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tokens.Parse(bad); err == nil {
				t.Fatal("invalid token accepted")
			}
		})
	}
}
func TestRefreshEntropyAndHash(t *testing.T) {
	a, err := NewRefresh()
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewRefresh()
	if err != nil {
		t.Fatal(err)
	}
	if a == b || !ValidRefresh(a) || len(Hash(a)) != 64 || Hash(a) == a || ValidRefresh("invalid") {
		t.Fatal("invalid refresh generation")
	}
}
