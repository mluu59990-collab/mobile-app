package storage

import (
	"context"
	"testing"
	"time"
)

func TestPresignPutSignsContentLengthAndContentType(t *testing.T) {
	client, err := NewR2(context.Background(), R2Config{
		AccountID: "account-id", AccessKey: "test-access", SecretKey: "test-secret", Bucket: "private-images",
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.PresignPut(context.Background(), "tmp/users/user/image.jpg", "image/jpeg", 12345, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if result.URL == "" {
		t.Fatal("expected presigned URL")
	}
	if result.Headers["Content-Length"] != "12345" || result.Headers["Content-Type"] != "image/jpeg" {
		t.Fatalf("Content-Length and Content-Type were not both signed: headers=%v", result.Headers)
	}
}
