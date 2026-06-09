package s3infra

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const presignTTL = 15 * time.Minute

// Storage provides S3-compatible object storage for resume PDFs.
// All buckets are private — no public access. Pre-signed URLs only.
type Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

func New(ctx context.Context, endpoint, region, bucket, accessKey, secretKey string) (*Storage, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	}

	// Use custom endpoint for MinIO / local dev
	customEndpoint := endpoint
	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("s3: load config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if customEndpoint != "" {
			o.BaseEndpoint = aws.String(customEndpoint)
			o.UsePathStyle = true // required for MinIO
		}
	})

	return &Storage{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		bucket:        bucket,
	}, nil
}

// Upload stores an object in S3. contentLength is used for progress tracking; pass -1 if unknown.
func (s *Storage) Upload(ctx context.Context, key string, body io.Reader, contentLength int64) error {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String("application/pdf"),
	}
	if contentLength > 0 {
		input.ContentLength = aws.Int64(contentLength)
	}
	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("s3: upload %s: %w", key, err)
	}
	return nil
}

// GetPresignedURL generates a short-lived pre-signed GET URL.
// TTL is fixed at 15 minutes — never expose raw S3 URLs.
func (s *Storage) GetPresignedURL(ctx context.Context, key string) (string, error) {
	req, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = presignTTL
	})
	if err != nil {
		return "", fmt.Errorf("s3: presign %s: %w", key, err)
	}
	return req.URL, nil
}

// Delete removes an object from S3.
func (s *Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// ResumeKey generates a consistent S3 key for a resume PDF.
// Format: resumes/{candidate_id}/{resume_id}.pdf
func ResumeKey(candidateID, resumeID string) string {
	return fmt.Sprintf("resumes/%s/%s.pdf", candidateID, resumeID)
}
