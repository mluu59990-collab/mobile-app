package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"strings"
	"time"

	"expense-app/backend/internal/storage"
	"github.com/google/uuid"
)

const maxExpenseImageSize int64 = 10 << 20

var (
	ErrInvalidImage            = errors.New("invalid image")
	ErrInvalidImageKey         = errors.New("invalid image key")
	ErrImageStorageUnavailable = errors.New("image storage unavailable")
)

type ExpenseImageConfig struct {
	PutURLTTL time.Duration
	GetURLTTL time.Duration
}

type PresignExpenseImageResult struct {
	UploadURL string            `json:"upload_url"`
	ImageKey  string            `json:"image_key"`
	ExpiresIn int64             `json:"expires_in"`
	Headers   map[string]string `json:"headers"`
}

type ImageStorage interface {
	PresignPut(context.Context, string, string, int64, time.Duration) (storage.PresignedPut, error)
	Head(context.Context, string) (storage.ObjectMetadata, error)
	ReadPrefix(context.Context, string, int64) ([]byte, error)
	Copy(context.Context, string, string) error
	PresignGet(context.Context, string, time.Duration) (string, error)
	Delete(context.Context, string) error
}

func (s *ExpenseService) PresignImageUpload(ctx context.Context, userID, contentType string, size int64) (PresignExpenseImageResult, error) {
	ext, ok := imageExtensions[contentType]
	if !ok || size <= 0 || size > maxExpenseImageSize {
		return PresignExpenseImageResult{}, ErrInvalidImage
	}
	if s.images == nil {
		return PresignExpenseImageResult{}, ErrImageStorageUnavailable
	}
	key := path.Join("tmp", "users", userID, uuid.NewString()+ext)
	put, err := s.images.PresignPut(ctx, key, contentType, size, s.imageConfig.PutURLTTL)
	if err != nil {
		return PresignExpenseImageResult{}, err
	}
	return PresignExpenseImageResult{
		UploadURL: put.URL, ImageKey: key, ExpiresIn: int64(s.imageConfig.PutURLTTL.Seconds()), Headers: put.Headers,
	}, nil
}

func (s *ExpenseService) FinalizeImage(ctx context.Context, userID, temporaryKey string) (string, error) {
	if s.images == nil {
		return "", ErrImageStorageUnavailable
	}
	contentType, ext, err := validateTemporaryImageKey(userID, temporaryKey)
	if err != nil {
		return "", err
	}
	metadata, err := s.images.Head(ctx, temporaryKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return "", ErrInvalidImageKey
		}
		return "", err
	}
	if err := validateImageMetadata(metadata, contentType); err != nil {
		return "", err
	}
	permanentKey := path.Join("users", userID, uuid.NewString()+ext)
	if err := s.images.Copy(ctx, temporaryKey, permanentKey); err != nil {
		return "", err
	}
	if err := s.verifyImage(ctx, permanentKey, contentType); err != nil {
		s.cleanupImage(ctx, "remove_invalid_copy", permanentKey)
		return "", err
	}
	s.cleanupImage(ctx, "remove_temporary_upload", temporaryKey)
	return permanentKey, nil
}

func (s *ExpenseService) verifyImage(ctx context.Context, key, expectedType string) error {
	metadata, err := s.images.Head(ctx, key)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return ErrInvalidImageKey
		}
		return err
	}
	if err := validateImageMetadata(metadata, expectedType); err != nil {
		return err
	}
	prefix, err := s.images.ReadPrefix(ctx, key, 512)
	if err != nil || detectImageType(prefix) != expectedType {
		return ErrInvalidImage
	}
	return nil
}

func validateImageMetadata(metadata storage.ObjectMetadata, expectedType string) error {
	if _, ok := imageExtensions[metadata.ContentType]; !ok || metadata.ContentType != expectedType || metadata.Size <= 0 || metadata.Size > maxExpenseImageSize {
		return ErrInvalidImage
	}
	return nil
}

var imageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func validateTemporaryImageKey(userID, key string) (string, string, error) {
	parts := strings.Split(key, "/")
	if len(parts) != 4 || parts[0] != "tmp" || parts[1] != "users" || parts[2] != userID {
		return "", "", ErrInvalidImageKey
	}
	filename := parts[3]
	for contentType, ext := range imageExtensions {
		if strings.HasSuffix(filename, ext) {
			id := strings.TrimSuffix(filename, ext)
			if parsed, err := uuid.Parse(id); err == nil && parsed.String() == id {
				return contentType, ext, nil
			}
		}
	}
	return "", "", ErrInvalidImageKey
}

func detectImageType(prefix []byte) string {
	switch {
	case len(prefix) >= 3 && prefix[0] == 0xff && prefix[1] == 0xd8 && prefix[2] == 0xff:
		return "image/jpeg"
	case len(prefix) >= 8 && string(prefix[:8]) == "\x89PNG\r\n\x1a\n":
		return "image/png"
	case len(prefix) >= 12 && string(prefix[:4]) == "RIFF" && string(prefix[8:12]) == "WEBP":
		return "image/webp"
	default:
		return ""
	}
}

func (s *ExpenseService) SignedImageURL(ctx context.Context, imageKey *string) (*string, error) {
	if imageKey == nil || *imageKey == "" {
		return nil, nil
	}
	if s.images == nil {
		return nil, ErrImageStorageUnavailable
	}
	url, err := s.images.PresignGet(ctx, *imageKey, s.imageConfig.GetURLTTL)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

func (s *ExpenseService) cleanupImage(ctx context.Context, operation, key string) {
	if err := s.images.Delete(ctx, key); err != nil {
		slog.ErrorContext(ctx, "expense image cleanup failed", "operation", operation, "error_type", fmt.Sprintf("%T", err))
	}
}
