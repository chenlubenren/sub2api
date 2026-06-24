package service

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3StorageProvider struct {
	client    *s3.Client
	presigner *s3.PresignClient
}

func NewS3StorageProvider(ctx context.Context, cfg *config.Config) (*S3StorageProvider, error) {
	if cfg == nil || cfg.Storage.Backend != config.StorageBackendS3 {
		return nil, ErrFileStorageNotConfigured
	}

	serverClient, err := newS3Client(ctx, cfg, cfg.Storage.Endpoint)
	if err != nil {
		return nil, err
	}
	presignEndpoint := strings.TrimSpace(cfg.Storage.PublicEndpoint)
	if presignEndpoint == "" {
		presignEndpoint = cfg.Storage.Endpoint
	}
	presignClient, err := newS3Client(ctx, cfg, presignEndpoint)
	if err != nil {
		return nil, err
	}

	return &S3StorageProvider{
		client:    serverClient,
		presigner: s3.NewPresignClient(presignClient),
	}, nil
}

func (p *S3StorageProvider) PresignUpload(ctx context.Context, input PresignUploadInput) (PresignedUploadInfo, error) {
	if p == nil || p.presigner == nil {
		return PresignedUploadInfo{}, ErrFileStorageNotConfigured
	}
	expiry := input.Expires
	if expiry <= 0 {
		expiry = time.Duration(config.DefaultStoragePresignExpire) * time.Second
	}
	req, err := p.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(input.Bucket),
		Key:         aws.String(input.ObjectKey),
		ContentType: aws.String(input.ContentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return PresignedUploadInfo{}, err
	}
	headers := map[string]string{"Content-Type": input.ContentType}
	for key, values := range req.SignedHeader {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return PresignedUploadInfo{
		Method:    "PUT",
		URL:       req.URL,
		Headers:   headers,
		ExpiresAt: time.Now().Add(expiry),
	}, nil
}

func (p *S3StorageProvider) PresignDownload(ctx context.Context, file *FileObject, expires time.Duration) (string, error) {
	if p == nil || p.presigner == nil || file == nil {
		return "", ErrFileStorageNotConfigured
	}
	if expires <= 0 {
		expires = time.Duration(config.DefaultStoragePresignExpire) * time.Second
	}
	req, err := p.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(file.Bucket),
		Key:    aws.String(file.ObjectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (p *S3StorageProvider) VerifyUploaded(ctx context.Context, file *FileObject) error {
	if p == nil || p.client == nil || file == nil {
		return ErrFileStorageNotConfigured
	}
	_, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(file.Bucket),
		Key:    aws.String(file.ObjectKey),
	})
	return err
}

func newS3Client(ctx context.Context, cfg *config.Config, endpoint string) (*s3.Client, error) {
	region := strings.TrimSpace(cfg.Storage.Region)
	if region == "" {
		region = "us-east-1"
	}
	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}
	if cfg.Storage.AccessKey != "" || cfg.Storage.SecretKey != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
		))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, err
	}
	endpoint = strings.TrimSpace(endpoint)
	return s3.NewFromConfig(awsCfg, func(opts *s3.Options) {
		if endpoint != "" {
			opts.BaseEndpoint = aws.String(endpoint)
		}
		opts.UsePathStyle = cfg.Storage.UsePathStyle
	}), nil
}
