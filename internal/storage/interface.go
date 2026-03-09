package storage

import (
	"context"
	"io"
	"time"
)

type Client interface {
	Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, key string) error
}
