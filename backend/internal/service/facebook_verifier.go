package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const facebookGraphURL = "https://graph.facebook.com"

var errFacebookUnavailable = errors.New("Facebook Graph API unavailable")

type FacebookGraphVerifier struct {
	AppID     string
	AppSecret string
	Client    *http.Client
}

type facebookDebugResponse struct {
	Data struct {
		IsValid   bool   `json:"is_valid"`
		Type      string `json:"type"`
		AppID     string `json:"app_id"`
		UserID    string `json:"user_id"`
		ExpiresAt int64  `json:"expires_at"`
	} `json:"data"`
}

type facebookMeResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func (v FacebookGraphVerifier) Verify(ctx context.Context, userAccessToken string) (FacebookIdentity, error) {
	if v.AppID == "" || v.AppSecret == "" || userAccessToken == "" {
		return FacebookIdentity{}, fmt.Errorf("facebook: verifier is not configured")
	}
	client := v.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	clientCopy := *client
	clientCopy.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	client = &clientCopy

	debugURL, err := url.Parse(facebookGraphURL + "/debug_token")
	if err != nil {
		return FacebookIdentity{}, err
	}
	debugQuery := debugURL.Query()
	debugQuery.Set("input_token", userAccessToken)
	debugQuery.Set("access_token", v.AppID+"|"+v.AppSecret)
	debugURL.RawQuery = debugQuery.Encode()
	var debug facebookDebugResponse
	if err := v.getJSON(ctx, client, debugURL.String(), "", &debug); err != nil {
		if errors.Is(err, errFacebookUnavailable) {
			return FacebookIdentity{}, ErrFacebookUnavailable
		}
		return FacebookIdentity{}, err
	}
	if !debug.Data.IsValid || debug.Data.Type != "USER" || debug.Data.AppID != v.AppID || debug.Data.UserID == "" ||
		(debug.Data.ExpiresAt != 0 && debug.Data.ExpiresAt <= time.Now().Unix()) {
		return FacebookIdentity{}, fmt.Errorf("facebook: access token is invalid for this app")
	}

	meURL, err := url.Parse(facebookGraphURL + "/me")
	if err != nil {
		return FacebookIdentity{}, err
	}
	query := meURL.Query()
	query.Set("fields", "id,email")
	meURL.RawQuery = query.Encode()
	var me facebookMeResponse
	if err := v.getJSON(ctx, client, meURL.String(), userAccessToken, &me); err != nil {
		if errors.Is(err, errFacebookUnavailable) {
			return FacebookIdentity{}, ErrFacebookUnavailable
		}
		return FacebookIdentity{}, err
	}
	if me.ID == "" || me.ID != debug.Data.UserID {
		return FacebookIdentity{}, fmt.Errorf("facebook: profile does not match inspected token")
	}
	var email *string
	if strings.TrimSpace(me.Email) != "" {
		value := me.Email
		email = &value
	}
	return FacebookIdentity{Subject: me.ID, Email: email}, nil
}

func (v FacebookGraphVerifier) getJSON(ctx context.Context, client *http.Client, endpoint, bearer string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: request failed", errFacebookUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("%w: HTTP %d", errFacebookUnavailable, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("facebook: provider rejected request")
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 16*1024)).Decode(target); err != nil {
		return fmt.Errorf("%w: invalid provider response", errFacebookUnavailable)
	}
	return nil
}
