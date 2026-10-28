package repository

import (
	"context"
	"errors"
	"time"

	"expense-app/backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Auth struct{ db *gorm.DB }

func NewAuth(db *gorm.DB) *Auth { return &Auth{db: db} }

func (r *Auth) Register(ctx context.Context, user *model.User, token *model.RefreshToken) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		identity := model.AuthIdentity{ID: uuid.NewString(), UserID: user.ID, Provider: "password", ProviderUserID: *user.Email}
		if err := tx.Create(&identity).Error; err != nil {
			return err
		}
		return tx.Create(token).Error
	})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "auth_identities_provider_provider_user_id_key" {
		return model.ErrEmailTaken
	}
	return err
}
func (r *Auth) PasswordUser(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Table("users").Select("users.*").Joins("JOIN auth_identities ON auth_identities.user_id = users.id").Where("auth_identities.provider = ? AND auth_identities.provider_user_id = ?", "password", email).Take(&user).Error
	return user, notFound(err)
}
func (r *Auth) User(ctx context.Context, id string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&user).Error
	return user, notFound(err)
}
func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrNotFound
	}
	return err
}
func (r *Auth) SaveRefresh(ctx context.Context, token *model.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// The row lock serializes refresh and logout. Any failure rolls back revocation.
func (r *Auth) Rotate(ctx context.Context, oldHash string, next *model.RefreshToken) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var old model.RefreshToken
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token_hash = ?", oldHash).Take(&old).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.ErrRefreshInvalid
		}
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if old.RevokedAt != nil || !old.ExpiresAt.After(now) {
			return model.ErrRefreshInvalid
		}
		if err := tx.Where("id = ?", old.UserID).Take(&user).Error; err != nil {
			return err
		}
		if err := tx.Model(&old).Update("revoked_at", now).Error; err != nil {
			return err
		}
		next.UserID = old.UserID
		return tx.Create(next).Error
	})
	return user, err
}
func (r *Auth) Revoke(ctx context.Context, hash string) error {
	return r.db.WithContext(ctx).Model(&model.RefreshToken{}).Where("token_hash = ? AND revoked_at IS NULL", hash).Update("revoked_at", time.Now().UTC()).Error
}

func (r *Auth) GoogleUser(ctx context.Context, subject string, email *string, verifiedAuthoritative bool) (model.User, error) {
	return r.federatedUser(ctx, "google", subject, email, verifiedAuthoritative)
}

func (r *Auth) FacebookUser(ctx context.Context, subject string, email *string) (model.User, error) {
	// Facebook's /me email field does not prove that the provider verified
	// control of that address. Store it when available but never auto-link it.
	return r.federatedUser(ctx, "facebook", subject, email, false)
}

func (r *Auth) federatedUser(ctx context.Context, provider, subject string, email *string, verifiedAuthoritative bool) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var identity model.AuthIdentity
		err := tx.Where("provider = ? AND provider_user_id = ?", provider, subject).Take(&identity).Error
		if err == nil {
			return tx.Where("id = ?", identity.UserID).Take(&user).Error
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if email != nil && verifiedAuthoritative {
			err := tx.Where("email = ?", *email).Take(&user).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if user.ID == "" {
			user = model.User{ID: uuid.NewString(), Email: email}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
		}
		return tx.Create(&model.AuthIdentity{ID: uuid.NewString(), UserID: user.ID, Provider: provider, ProviderUserID: subject}).Error
	})
	if err == nil {
		return user, nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "auth_identities_provider_provider_user_id_key" {
		// Concurrent first sign-ins for the same Google subject: the losing
		// transaction rolls back its user and reads the identity created by the winner.
		var identity model.AuthIdentity
		if lookupErr := r.db.WithContext(ctx).Where("provider = ? AND provider_user_id = ?", provider, subject).Take(&identity).Error; lookupErr == nil {
			if lookupErr = r.db.WithContext(ctx).Where("id = ?", identity.UserID).Take(&user).Error; lookupErr == nil {
				return user, nil
			}
		}
	}
	return model.User{}, err
}
