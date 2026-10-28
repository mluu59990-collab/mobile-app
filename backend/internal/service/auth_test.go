package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"expense-app/backend/internal/auth"
	"expense-app/backend/internal/config"
	"expense-app/backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

// Stub only external persistence; password and token logic run unchanged.
type stubRepo struct {
	AuthRepository
	user            model.User
	token           *model.RefreshToken
	err             error
	googleSubject   string
	googleEmail     *string
	googleLinked    bool
	facebookSubject string
	facebookEmail   *string
}

func (r *stubRepo) Register(_ context.Context, u *model.User, token *model.RefreshToken) error {
	r.user = *u
	r.token = token
	return r.err
}
func (r *stubRepo) PasswordUser(context.Context, string) (model.User, error) { return r.user, r.err }
func (r *stubRepo) SaveRefresh(_ context.Context, token *model.RefreshToken) error {
	r.token = token
	return r.err
}
func (r *stubRepo) GoogleUser(_ context.Context, subject string, email *string, verifiedAuthoritative bool) (model.User, error) {
	r.googleSubject, r.googleEmail, r.googleLinked = subject, email, verifiedAuthoritative
	return r.user, r.err
}
func (r *stubRepo) FacebookUser(_ context.Context, subject string, email *string) (model.User, error) {
	r.facebookSubject, r.facebookEmail = subject, email
	return r.user, r.err
}

type googleVerifierStub struct {
	identity GoogleIdentity
	audience string
	err      error
}

func (v *googleVerifierStub) Verify(_ context.Context, raw, audience string) (GoogleIdentity, error) {
	v.audience = audience
	if raw != "fake-google-id-token" {
		return GoogleIdentity{}, errors.New("invalid token")
	}
	return v.identity, v.err
}

type facebookVerifierStub struct {
	identity FacebookIdentity
	err      error
}

func (v facebookVerifierStub) Verify(_ context.Context, token string) (FacebookIdentity, error) {
	if token != "fake-facebook-access-token" {
		return FacebookIdentity{}, errors.New("invalid token")
	}
	return v.identity, v.err
}
func TestCredentialsValidation(t *testing.T) {
	for _, input := range []struct{ email, password string }{{"bad", "password123"}, {"Name <a@example.com>", "password123"}, {"a@example.com", "short"}, {"a@example.com", strings.Repeat("x", 73)}} {
		if _, err := credentials(input.email, input.password); !errors.Is(err, ErrInvalidInput) {
			t.Fatal("invalid input accepted")
		}
	}
	if got, err := credentials("  TEST@example.com ", "password123"); err != nil || got != "test@example.com" {
		t.Fatal("normalization failed")
	}
}
func TestRegisterAndLogin(t *testing.T) {
	repo := &stubRepo{}
	svc, err := NewAuth(repo, auth.Tokens{Config: config.Auth{Secret: strings.Repeat("s", 32), Issuer: "test", Audience: "test", AccessTTL: time.Minute, RefreshTTL: time.Hour}})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	session, err := svc.Register(ctx, " TEST@example.com ", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if *repo.user.Email != "test@example.com" || *repo.user.PasswordHash == "password123" || repo.token.TokenHash != auth.Hash(session.RefreshToken) {
		t.Fatal("unsafe stored credentials")
	}
	if cost, err := bcrypt.Cost([]byte(*repo.user.PasswordHash)); err != nil || cost < 12 {
		t.Fatal("weak password hash")
	}
	if _, err := svc.Login(ctx, "test@example.com", "password123"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Login(ctx, "test@example.com", "wrong-password"); !errors.Is(err, ErrCredentials) {
		t.Fatal("wrong password accepted")
	}
	repo.err = model.ErrNotFound
	if _, err := svc.Login(ctx, "missing@example.com", "password123"); !errors.Is(err, ErrCredentials) {
		t.Fatal("missing user leaked")
	}
	repo.err = errors.New("db unavailable")
	if _, err := svc.Login(ctx, "test@example.com", "password123"); !errors.Is(err, repo.err) {
		t.Fatal("DB failure swallowed")
	}
}

func TestGoogleLogin(t *testing.T) {
	email := "person@gmail.com"
	repo := &stubRepo{user: model.User{ID: "google-user", Email: &email}}
	verifier := &googleVerifierStub{identity: GoogleIdentity{Subject: "google-subject-123", Email: email, EmailVerified: true}}
	svc, err := NewAuth(repo, auth.Tokens{Config: config.Auth{Secret: strings.Repeat("s", 32), Issuer: "test", Audience: "mobile", GoogleClientID: "google-client", AccessTTL: time.Minute, RefreshTTL: time.Hour}}, verifier)
	if err != nil {
		t.Fatal(err)
	}
	session, err := svc.GoogleLogin(context.Background(), "fake-google-id-token")
	if err != nil {
		t.Fatal(err)
	}
	if verifier.audience != "google-client" || repo.googleSubject != "google-subject-123" || !repo.googleLinked || repo.googleEmail == nil || *repo.googleEmail != email {
		t.Fatal("verified Google identity was not verified/linked using subject and configured audience")
	}
	if session.User.ID != repo.user.ID || session.AccessToken == "" || session.RefreshToken == "" || repo.token.TokenHash != auth.Hash(session.RefreshToken) {
		t.Fatal("Google login did not create a first-party session")
	}

	verifier.identity = GoogleIdentity{Subject: "google-subject-456", Email: "person@example.net", EmailVerified: true}
	if _, err := svc.GoogleLogin(context.Background(), "fake-google-id-token"); err != nil {
		t.Fatal(err)
	}
	if repo.googleLinked {
		t.Fatal("untrusted external email domain was linked automatically")
	}

	verifier.err = errors.New("signature invalid")
	if _, err := svc.GoogleLogin(context.Background(), "fake-google-id-token"); !errors.Is(err, ErrGoogleToken) {
		t.Fatal("invalid Google token accepted")
	}
	if _, err := svc.GoogleLogin(context.Background(), "untrusted"); !errors.Is(err, ErrGoogleToken) {
		t.Fatal("unverified token accepted")
	}
}

func TestGoogleLoginRequiresConfiguredClient(t *testing.T) {
	svc, err := NewAuth(&stubRepo{}, auth.Tokens{Config: config.Auth{Secret: strings.Repeat("s", 32)}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GoogleLogin(context.Background(), "fake-google-id-token"); !errors.Is(err, ErrGoogleConfig) {
		t.Fatal("Google login should require a client ID")
	}
}

func TestFacebookLoginAllowsMissingEmail(t *testing.T) {
	repo := &stubRepo{user: model.User{ID: "facebook-user"}}
	verifier := facebookVerifierStub{identity: FacebookIdentity{Subject: "facebook-subject"}}
	svc, err := NewAuth(repo, auth.Tokens{Config: config.Auth{Secret: strings.Repeat("s", 32), FacebookAppID: "app-id", FacebookAppSecret: "app-secret", AccessTTL: time.Minute, RefreshTTL: time.Hour}})
	if err != nil {
		t.Fatal(err)
	}
	svc.SetFacebookVerifier(verifier)
	session, err := svc.FacebookLogin(context.Background(), "fake-facebook-access-token")
	if err != nil {
		t.Fatal(err)
	}
	if repo.facebookSubject != "facebook-subject" || repo.facebookEmail != nil || session.User.Email != nil || session.AccessToken == "" || session.RefreshToken == "" {
		t.Fatal("Facebook login should issue an app session without requiring email")
	}

	value := "FB.USER@example.com"
	verifier.identity = FacebookIdentity{Subject: "facebook-subject-with-email", Email: &value}
	svc.SetFacebookVerifier(verifier)
	if _, err := svc.FacebookLogin(context.Background(), "fake-facebook-access-token"); err != nil {
		t.Fatal(err)
	}
	if repo.facebookEmail == nil || *repo.facebookEmail != "fb.user@example.com" {
		t.Fatal("Facebook email was not normalized")
	}
	if _, err := svc.FacebookLogin(context.Background(), "invalid"); !errors.Is(err, ErrFacebookToken) {
		t.Fatal("invalid Facebook token accepted")
	}
}
