package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type S3Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	UseSSL          bool
}

type S3Client struct {
	client *s3.Client
	config S3Config
}

func NewS3Client(cfg S3Config) (*S3Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	clientOptions := []func(*s3.Options){}

	if cfg.Endpoint != "" {
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, clientOptions...)

	return &S3Client{
		client: client,
		config: cfg,
	}, nil
}

func (s *S3Client) UploadFile(ctx context.Context, file io.Reader, filename string, contentType string) (*string, error) {
	key := s.generateKey(filename)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.config.Bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	fileURL := s.getPublicURL(key)
	return &fileURL, nil
}

func (s *S3Client) DeleteFile(ctx context.Context, fileURL string) error {
	key, err := s.extractKeyFromURL(fileURL)
	if err != nil {
		return fmt.Errorf("invalid file URL: %w", err)
	}

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (s *S3Client) GetPresignedURL(ctx context.Context, fileURL string, expiresIn time.Duration) (*string, error) {
	key, err := s.extractKeyFromURL(fileURL)
	if err != nil {
		return nil, fmt.Errorf("invalid file URL: %w", err)
	}

	presignedClient := s3.NewPresignClient(s.client)
	request, err := presignedClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.config.Bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &request.URL, nil
}

func (s *S3Client) generateKey(filename string) string {
	uniqueID := uuid.New().String()
	timestamp := time.Now().Format("2006-01-02")
	return fmt.Sprintf("uploads/%s/%s-%s", timestamp, uniqueID, filename)
}

func (s *S3Client) getPublicURL(key string) string {
	scheme := "https"
	if !s.config.UseSSL {
		scheme = "http"
	}

	if s.config.Endpoint != "" {
		endpointURL, err := url.Parse(s.config.Endpoint)
		if err == nil {
			return fmt.Sprintf("%s://%s/%s/%s", scheme, endpointURL.Host, s.config.Bucket, key)
		}
	}

	return fmt.Sprintf("https://%s.%s.amazonaws.com/%s", s.config.Bucket, s.config.Region, key)
}

func (s *S3Client) extractKeyFromURL(fileURL string) (string, error) {
	u, err := url.Parse(fileURL)
	if err != nil {
		return "", err
	}

	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid URL format")
	}

	return parts[1], nil
}