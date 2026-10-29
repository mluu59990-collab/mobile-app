package storage

import (
	"context"
	"errors"
	"time"
)

var ErrObjectNotFound = errors.New("object not found")

type PresignedPut struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type ObjectMetadata struct {
	Size        int64
	ContentType string
}

type ImageStorage interface {
	PresignPut(context.Context, string, string, int64, time.Duration) (PresignedPut, error)
	Head(context.Context, string) (ObjectMetadata, error)
	ReadPrefix(context.Context, string, int64) ([]byte, error)
	Copy(context.Context, string, string) error
	PresignGet(context.Context, string, time.Duration) (string, error)
	Delete(context.Context, string) error
}
