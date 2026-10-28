package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"expense-app/backend/internal/auth"
	"expense-app/backend/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrCredentials         = errors.New("invalid credentials")
	ErrGoogleToken         = errors.New("invalid Google ID token")
	ErrGoogleConfig        = errors.New("Google login is not configured")
	ErrFacebookToken       = errors.New("invalid Facebook access token")
	ErrFacebookConfig      = errors.New("Facebook login is not configured")
	ErrFacebookUnavailable = errors.New("Facebook verification is unavailable")
)

type AuthRepository interface {
	Register(context.Context, *model.User, *model.RefreshToken) error
	PasswordUser(context.Context, string) (model.User, error)
	User(context.Context, string) (model.User, error)
	SaveRefresh(context.Context, *model.RefreshToken) error
	Rotate(context.Context, string, *model.RefreshToken) (model.User, error)
	Revoke(context.Context, string) error
	GoogleUser(context.Context, string, *string, bool) (model.User, error)
	FacebookUser(context.Context, string, *string) (model.User, error)
}

type GoogleIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	HostedDomain  string
}

type GoogleVerifier interface {
	Verify(context.Context, string, string) (GoogleIdentity, error)
}

type FacebookIdentity struct {
	Subject string
	Email   *string
}

type FacebookVerifier interface {
	Verify(context.Context, string) (FacebookIdentity, error)
}

type Auth struct {
	repo      AuthRepository
	tokens    auth.Tokens
	dummyHash []byte
	google    GoogleVerifier
	facebook  FacebookVerifier
}

func (s *Auth) SetFacebookVerifier(verifier FacebookVerifier) {
	s.facebook = verifier
}

type UserView struct {
	ID    string  `json:"id"`
	Email *string `json:"email"`
}
type Session struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserView `json:"user"`
}

func NewAuth(repo AuthRepository, tokens auth.Tokens, google ...GoogleVerifier) (*Auth, error) {
	dummy, err := bcrypt.GenerateFromPassword([]byte(uuid.NewString()), 12)
	if err != nil {
		return nil, err
	}
	service := &Auth{repo: repo, tokens: tokens, dummyHash: dummy}
	if len(google) > 0 {
		service.google = google[0]
	}
	return service, nil
}

func (s *Auth) GoogleLogin(ctx context.Context, idToken string) (Session, error) {
	if s.google == nil || s.tokens.Config.GoogleClientID == "" {
		return Session{}, ErrGoogleConfig
	}
	if len(idToken) == 0 || len(idToken) > 8192 {
		return Session{}, ErrGoogleToken
	}
	identity, err := s.google.Verify(ctx, idToken, s.tokens.Config.GoogleClientID)
	if err != nil || identity.Subject == "" {
		return Session{}, ErrGoogleToken
	}
	var email *string
	if identity.EmailVerified && identity.Email != "" {
		normalized := strings.ToLower(strings.TrimSpace(identity.Email))
		if parsed, parseErr := mail.ParseAddress(normalized); parseErr == nil && parsed.Address == normalized && len(normalized) <= 254 {
			email = &normalized
		}
	}
	user, err := s.repo.GoogleUser(ctx, identity.Subject, email, identity.EmailVerified && email != nil && googleAuthoritative(identity))
	if err != nil {
		return Session{}, err
	}
	raw, token, err := s.refreshRecord()
	if err != nil {
		return Session{}, err
	}
	token.UserID = user.ID
	result, err := s.session(user, raw)
	if err != nil {
		return Session{}, err
	}
	if err := s.repo.SaveRefresh(ctx, token); err != nil {
		return Session{}, err
	}
	return result, nil
}

func googleAuthoritative(identity GoogleIdentity) bool {
	email := strings.ToLower(identity.Email)
	return strings.HasSuffix(email, "@gmail.com") || (identity.EmailVerified && identity.HostedDomain != "")
}

func (s *Auth) FacebookLogin(ctx context.Context, accessToken string) (Session, error) {
	if s.facebook == nil || s.tokens.Config.FacebookAppID == "" || s.tokens.Config.FacebookAppSecret == "" {
		return Session{}, ErrFacebookConfig
	}
	if len(accessToken) == 0 || len(accessToken) > 4096 {
		return Session{}, ErrFacebookToken
	}
	identity, err := s.facebook.Verify(ctx, accessToken)
	if errors.Is(err, ErrFacebookUnavailable) {
		return Session{}, ErrFacebookUnavailable
	}
	if err != nil || identity.Subject == "" {
		return Session{}, ErrFacebookToken
	}
	var email *string
	if identity.Email != nil {
		normalized := strings.ToLower(strings.TrimSpace(*identity.Email))
		if parsed, parseErr := mail.ParseAddress(normalized); parseErr == nil && parsed.Address == normalized && len(normalized) <= 254 {
			email = &normalized
		}
	}
	user, err := s.repo.FacebookUser(ctx, identity.Subject, email)
	if err != nil {
		return Session{}, err
	}
	raw, token, err := s.refreshRecord()
	if err != nil {
		return Session{}, err
	}
	token.UserID = user.ID
	result, err := s.session(user, raw)
	if err != nil {
		return Session{}, err
	}
	if err := s.repo.SaveRefresh(ctx, token); err != nil {
		return Session{}, err
	}
	return result, nil
}
func credentials(email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || len(email) > 254 || len(password) < 8 || len(password) > 72 {
		return "", ErrInvalidInput
	}
	return email, nil
}
func (s *Auth) refreshRecord() (string, *model.RefreshToken, error) {
	raw, err := auth.NewRefresh()
	if err != nil {
		return "", nil, err
	}
	return raw, &model.RefreshToken{ID: uuid.NewString(), TokenHash: auth.Hash(raw), ExpiresAt: time.Now().UTC().Add(s.tokens.Config.RefreshTTL)}, nil
}
func (s *Auth) session(user model.User, raw string) (Session, error) {
	access, err := s.tokens.Access(user.ID)
	if err != nil {
		return Session{}, err
	}
	return Session{AccessToken: access, RefreshToken: raw, TokenType: "Bearer", ExpiresIn: int64(s.tokens.Config.AccessTTL / time.Second), User: UserView{ID: user.ID, Email: user.Email}}, nil
}
func (s *Auth) Register(ctx context.Context, email, password string) (Session, error) {
	email, err := credentials(email, password)
	if err != nil {
		return Session{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return Session{}, err
	}
	hashString := string(hash)
	user := model.User{ID: uuid.NewString(), Email: &email, PasswordHash: &hashString}
	raw, token, err := s.refreshRecord()
	if err != nil {
		return Session{}, err
	}
	token.UserID = user.ID
	result, err := s.session(user, raw)
	if err != nil {
		return Session{}, err
	}
	if err := s.repo.Register(ctx, &user, token); err != nil {
		return Session{}, err
	}
	return result, nil
}
func (s *Auth) Login(ctx context.Context, email, password string) (Session, error) {
	email, err := credentials(email, password)
	if err != nil {
		return Session{}, err
	}
	user, err := s.repo.PasswordUser(ctx, email)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return Session{}, err
	}
	hash := s.dummyHash
	if err == nil && user.PasswordHash != nil {
		hash = []byte(*user.PasswordHash)
	}
	compareErr := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if compareErr != nil || err != nil || user.PasswordHash == nil {
		return Session{}, ErrCredentials
	}
	raw, token, err := s.refreshRecord()
	if err != nil {
		return Session{}, err
	}
	token.UserID = user.ID
	result, err := s.session(user, raw)
	if err != nil {
		return Session{}, err
	}
	if err := s.repo.SaveRefresh(ctx, token); err != nil {
		return Session{}, err
	}
	return result, nil
}
func (s *Auth) Refresh(ctx context.Context, raw string) (Session, error) {
	if !auth.ValidRefresh(raw) {
		return Session{}, model.ErrRefreshInvalid
	}
	nextRaw, token, err := s.refreshRecord()
	if err != nil {
		return Session{}, err
	}
	user, err := s.repo.Rotate(ctx, auth.Hash(raw), token)
	if err != nil {
		return Session{}, err
	}
	return s.session(user, nextRaw)
}
func (s *Auth) Logout(ctx context.Context, raw string) error {
	if !auth.ValidRefresh(raw) {
		return fmt.Errorf("%w: malformed refresh token", ErrInvalidInput)
	}
	return s.repo.Revoke(ctx, auth.Hash(raw))
}
func (s *Auth) Me(ctx context.Context, id string) (UserView, error) {
	user, err := s.repo.User(ctx, id)
	return UserView{ID: user.ID, Email: user.Email}, err
}
