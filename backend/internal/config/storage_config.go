package config

import (
	"fmt"
	"strings"
)

type StorageConfig struct {
	Backend              string   `mapstructure:"backend"`
	Endpoint             string   `mapstructure:"endpoint"`
	PublicEndpoint       string   `mapstructure:"public_endpoint"`
	Region               string   `mapstructure:"region"`
	Bucket               string   `mapstructure:"bucket"`
	AccessKey            string   `mapstructure:"access_key"`
	SecretKey            string   `mapstructure:"secret_key"`
	UsePathStyle         bool     `mapstructure:"use_path_style"`
	PresignExpireSeconds int      `mapstructure:"presign_expire_seconds"`
	MaxFileSizeBytes     int64    `mapstructure:"max_file_size_bytes"`
	AllowedMimeTypes     []string `mapstructure:"allowed_mime_types"`
}

func normalizeStorageConfig(cfg *Config) {
	cfg.Storage.Backend = strings.ToLower(strings.TrimSpace(cfg.Storage.Backend))
	cfg.Storage.Endpoint = strings.TrimSpace(cfg.Storage.Endpoint)
	cfg.Storage.PublicEndpoint = strings.TrimSpace(cfg.Storage.PublicEndpoint)
	cfg.Storage.Region = strings.TrimSpace(cfg.Storage.Region)
	cfg.Storage.Bucket = strings.TrimSpace(cfg.Storage.Bucket)
	cfg.Storage.AccessKey = strings.TrimSpace(cfg.Storage.AccessKey)
	cfg.Storage.SecretKey = strings.TrimSpace(cfg.Storage.SecretKey)
	cfg.Storage.AllowedMimeTypes = normalizeLowerStringSlice(cfg.Storage.AllowedMimeTypes)
}

func (c *Config) validateStorageConfig() error {
	switch c.Storage.Backend {
	case "", StorageBackendDisabled:
		c.Storage.Backend = StorageBackendDisabled
		return nil
	case StorageBackendS3:
	default:
		return fmt.Errorf("storage.backend must be one of disabled/s3")
	}

	if c.Storage.Bucket == "" {
		return fmt.Errorf("storage.bucket is required when storage.backend=s3")
	}
	if c.Storage.Endpoint != "" {
		if err := ValidateAbsoluteHTTPURL(c.Storage.Endpoint); err != nil {
			return fmt.Errorf("storage.endpoint invalid: %w", err)
		}
		warnIfInsecureURL("storage.endpoint", c.Storage.Endpoint)
	}
	if c.Storage.PublicEndpoint != "" {
		if err := ValidateAbsoluteHTTPURL(c.Storage.PublicEndpoint); err != nil {
			return fmt.Errorf("storage.public_endpoint invalid: %w", err)
		}
		warnIfInsecureURL("storage.public_endpoint", c.Storage.PublicEndpoint)
	}
	if c.Storage.PresignExpireSeconds <= 0 {
		return fmt.Errorf("storage.presign_expire_seconds must be positive")
	}
	if c.Storage.MaxFileSizeBytes <= 0 {
		return fmt.Errorf("storage.max_file_size_bytes must be positive")
	}
	if len(c.Storage.AllowedMimeTypes) == 0 {
		return fmt.Errorf("storage.allowed_mime_types must not be empty when storage.backend=s3")
	}

	return nil
}
