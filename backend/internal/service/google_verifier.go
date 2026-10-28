package service

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

type GoogleIDTokenVerifier struct{}

func (GoogleIDTokenVerifier) Verify(ctx context.Context, raw, audience string) (GoogleIdentity, error) {
	if audience == "" {
		return GoogleIdentity{}, fmt.Errorf("Google client ID is not configured")
	}
	payload, err := idtoken.Validate(ctx, raw, audience)
	if err != nil {
		return GoogleIdentity{}, err
	}
	if (payload.Issuer != "accounts.google.com" && payload.Issuer != "https://accounts.google.com") || payload.Subject == "" {
		return GoogleIdentity{}, fmt.Errorf("idtoken: unexpected issuer or missing subject")
	}
	email, _ := payload.Claims["email"].(string)
	verified, _ := payload.Claims["email_verified"].(bool)
	hostedDomain, _ := payload.Claims["hd"].(string)
	return GoogleIdentity{Subject: payload.Subject, Email: email, EmailVerified: verified, HostedDomain: hostedDomain}, nil
}
