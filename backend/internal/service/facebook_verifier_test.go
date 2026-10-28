package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestFacebookGraphVerifierValidatesAppAndSubjectWithoutRequiringEmail(t *testing.T) {
	var calls int
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		calls++
		switch request.URL.Path {
		case "/debug_token":
			if request.URL.Query().Get("input_token") != "fake-access-token" || request.URL.Query().Get("access_token") != "app-id|app-secret" {
				t.Fatal("debug_token request did not use the supplied token and app credentials")
			}
			return jsonResponse(`{"data":{"is_valid":true,"type":"USER","app_id":"app-id","user_id":"facebook-user-id","expires_at":4102444800}}`), nil
		case "/me":
			if request.Header.Get("Authorization") != "Bearer fake-access-token" || request.URL.Query().Get("fields") != "id,email" {
				t.Fatal("profile request did not use the verified user token")
			}
			return jsonResponse(`{"id":"facebook-user-id"}`), nil
		default:
			t.Fatalf("unexpected Graph API path %q", request.URL.Path)
			return nil, nil
		}
	})}
	identity, err := (FacebookGraphVerifier{AppID: "app-id", AppSecret: "app-secret", Client: client}).Verify(context.Background(), "fake-access-token")
	if err != nil {
		t.Fatal(err)
	}
	if identity.Subject != "facebook-user-id" || identity.Email != nil || calls != 2 {
		t.Fatalf("expected verified subject with nil email; got %+v after %d calls", identity, calls)
	}
}

func TestFacebookGraphVerifierRejectsMismatchedOrInvalidToken(t *testing.T) {
	for _, body := range []string{
		`{"data":{"is_valid":false,"type":"USER","app_id":"app-id","user_id":"user"}}`,
		`{"data":{"is_valid":true,"type":"USER","app_id":"other-app","user_id":"user"}}`,
		`{"data":{"is_valid":true,"type":"USER","app_id":"app-id","user_id":"other-user"}}`,
	} {
		client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if request.URL.Path == "/debug_token" {
				return jsonResponse(body), nil
			}
			return jsonResponse(`{"id":"user"}`), nil
		})}
		if _, err := (FacebookGraphVerifier{AppID: "app-id", AppSecret: "app-secret", Client: client}).Verify(context.Background(), "fake-access-token"); err == nil {
			t.Fatalf("accepted invalid debug response %s", body)
		}
	}
}

func jsonResponse(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}
