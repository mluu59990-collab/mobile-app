package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"expense-app/backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Tokens struct{ Config config.Auth }

func (t Tokens) Access(userID string) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject: userID, Issuer: t.Config.Issuer, Audience: jwt.ClaimStrings{t.Config.Audience},
		ID: uuid.NewString(), IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(t.Config.AccessTTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(t.Config.Secret))
}

func (t Tokens) Parse(raw string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) { return []byte(t.Config.Secret), nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer(t.Config.Issuer), jwt.WithAudience(t.Config.Audience), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil {
		return "", err
	}
	if _, err := uuid.Parse(claims.Subject); err != nil || claims.ID == "" || claims.IssuedAt == nil || claims.NotBefore == nil {
		return "", fmt.Errorf("invalid access claims")
	}
	return claims.Subject, nil
}

func NewRefresh() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func Hash(raw string) string { sum := sha256.Sum256([]byte(raw)); return hex.EncodeToString(sum[:]) }
func ValidRefresh(raw string) bool {
	b, err := base64.RawURLEncoding.DecodeString(raw)
	return err == nil && len(b) == 32 && base64.RawURLEncoding.EncodeToString(b) == raw
}
