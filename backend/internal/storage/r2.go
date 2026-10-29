package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

type R2Config struct {
	AccountID string
	AccessKey string
	SecretKey string
	Bucket    string
}

type R2 struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

func NewR2(ctx context.Context, cfg R2Config) (*R2, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load R2 client configuration: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String("https://" + cfg.AccountID + ".r2.cloudflarestorage.com")
	})
	return &R2{client: client, presign: s3.NewPresignClient(client), bucket: cfg.Bucket}, nil
}

func (r *R2) PresignPut(ctx context.Context, key, contentType string, size int64, expires time.Duration) (PresignedPut, error) {
	request, err := r.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucket), Key: aws.String(key), ContentType: aws.String(contentType), ContentLength: aws.Int64(size),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return PresignedPut{}, fmt.Errorf("presign R2 image upload: %w", err)
	}
	headers := make(map[string]string, len(request.SignedHeader))
	for name, values := range request.SignedHeader {
		if len(values) > 0 && !strings.EqualFold(name, "host") {
			headers[http.CanonicalHeaderKey(name)] = values[0]
		}
	}
	return PresignedPut{URL: request.URL, Headers: headers}, nil
}

func (r *R2) Head(ctx context.Context, key string) (ObjectMetadata, error) {
	result, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(key)})
	if err != nil {
		var responseError *smithyhttp.ResponseError
		if errors.As(err, &responseError) && responseError.HTTPStatusCode() == 404 {
			return ObjectMetadata{}, ErrObjectNotFound
		}
		return ObjectMetadata{}, fmt.Errorf("head R2 object: %w", err)
	}
	return ObjectMetadata{Size: aws.ToInt64(result.ContentLength), ContentType: aws.ToString(result.ContentType)}, nil
}

func (r *R2) ReadPrefix(ctx context.Context, key string, size int64) ([]byte, error) {
	end := size - 1
	if end < 0 {
		end = 0
	}
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket), Key: aws.String(key), Range: aws.String(fmt.Sprintf("bytes=0-%d", end)),
	})
	if err != nil {
		return nil, fmt.Errorf("read R2 object prefix: %w", err)
	}
	defer result.Body.Close()
	data, err := io.ReadAll(io.LimitReader(result.Body, size))
	if err != nil {
		return nil, fmt.Errorf("read R2 object prefix body: %w", err)
	}
	return data, nil
}

func (r *R2) Copy(ctx context.Context, source, destination string) error {
	copySource := url.PathEscape(r.bucket) + "/" + strings.ReplaceAll(url.PathEscape(source), "%2F", "/")
	_, err := r.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket: aws.String(r.bucket), Key: aws.String(destination), CopySource: aws.String(copySource),
	})
	if err != nil {
		return fmt.Errorf("copy R2 image: %w", err)
	}
	return nil
}

func (r *R2) PresignGet(ctx context.Context, key string, expires time.Duration) (string, error) {
	request, err := r.presign.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(key)}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("presign R2 image read: %w", err)
	}
	return request.URL, nil
}

func (r *R2) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("delete R2 image: %w", err)
	}
	return nil
}
