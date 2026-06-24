package service

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/google/uuid"
)

const (
	FileObjectStatusPending  = "pending"
	FileObjectStatusUploaded = "uploaded"
	FileObjectStatusFailed   = "failed"
	FileObjectStatusExpired  = "expired"
)

type FileObject struct {
	ID               int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	OwnerUserID      int64
	APIKeyID         *int64
	Purpose          string
	StorageProvider  string
	Bucket           string
	ObjectKey        string
	OriginalFilename *string
	MimeType         string
	SizeBytes        int64
	SHA256           *string
	Status           string
	Metadata         map[string]any
	UploadedAt       *time.Time
	ExpiresAt        *time.Time
}

type CreateUploadSessionInput struct {
	OwnerUserID      int64
	APIKeyID         *int64
	Purpose          string
	OriginalFilename string
	MimeType         string
	SizeBytes        int64
	SHA256           *string
	Metadata         map[string]any
}

type UploadSession struct {
	FileID int64
	File   *FileObject
	Upload PresignedUploadInfo
}

type PresignUploadInput struct {
	Bucket      string
	ObjectKey   string
	ContentType string
	Expires     time.Duration
}

type PresignedUploadInfo struct {
	Method    string
	URL       string
	Headers   map[string]string
	ExpiresAt time.Time
}

type FileRepository interface {
	Create(ctx context.Context, file *FileObject) (*FileObject, error)
	GetByID(ctx context.Context, id int64) (*FileObject, error)
	UpdateStatus(ctx context.Context, id int64, status string, uploadedAt *time.Time) (*FileObject, error)
}

type StorageProvider interface {
	PresignUpload(ctx context.Context, input PresignUploadInput) (PresignedUploadInfo, error)
	VerifyUploaded(ctx context.Context, file *FileObject) error
}

type FileService struct {
	repo    FileRepository
	storage StorageProvider
	cfg     *config.Config
	now     func() time.Time
}

func NewFileService(repo FileRepository, storage StorageProvider, cfg *config.Config) *FileService {
	return &FileService{
		repo:    repo,
		storage: storage,
		cfg:     cfg,
		now:     time.Now,
	}
}

func (s *FileService) CreateUploadSession(ctx context.Context, input CreateUploadSessionInput) (*UploadSession, error) {
	if s == nil || s.repo == nil || s.storage == nil || s.cfg == nil {
		return nil, ErrFileStorageNotConfigured
	}
	storageCfg := s.cfg.Storage
	if storageCfg.Backend != config.StorageBackendS3 || storageCfg.Bucket == "" {
		return nil, ErrFileStorageNotConfigured
	}
	if input.OwnerUserID <= 0 || strings.TrimSpace(input.MimeType) == "" || input.SizeBytes < 0 {
		return nil, ErrFileInvalidInput
	}
	if storageCfg.MaxFileSizeBytes > 0 && input.SizeBytes > storageCfg.MaxFileSizeBytes {
		return nil, ErrFileInvalidInput
	}
	mimeType := strings.ToLower(strings.TrimSpace(input.MimeType))
	if !isAllowedFileMIMEType(mimeType, storageCfg.AllowedMimeTypes) {
		return nil, ErrFileInvalidInput
	}

	expiry := time.Duration(storageCfg.PresignExpireSeconds) * time.Second
	if expiry <= 0 {
		return nil, ErrFileInvalidInput
	}
	expiresAt := s.now().Add(expiry)
	filename := cleanUploadFilename(input.OriginalFilename)
	purpose := strings.TrimSpace(input.Purpose)
	if purpose == "" {
		purpose = "vision_input"
	}

	file := &FileObject{
		OwnerUserID:      input.OwnerUserID,
		APIKeyID:         cloneFileInt64Ptr(input.APIKeyID),
		Purpose:          purpose,
		StorageProvider:  config.StorageBackendS3,
		Bucket:           storageCfg.Bucket,
		ObjectKey:        fmt.Sprintf("files/%d/%s/%s", input.OwnerUserID, uuid.NewString(), filename),
		OriginalFilename: &filename,
		MimeType:         mimeType,
		SizeBytes:        input.SizeBytes,
		SHA256:           cloneFileStringPtr(input.SHA256),
		Status:           FileObjectStatusPending,
		Metadata:         cloneMetadata(input.Metadata),
		ExpiresAt:        &expiresAt,
	}

	created, err := s.repo.Create(ctx, file)
	if err != nil {
		return nil, err
	}
	upload, err := s.storage.PresignUpload(ctx, PresignUploadInput{
		Bucket:      created.Bucket,
		ObjectKey:   created.ObjectKey,
		ContentType: created.MimeType,
		Expires:     expiry,
	})
	if err != nil {
		return nil, err
	}
	return &UploadSession{
		FileID: created.ID,
		File:   created,
		Upload: upload,
	}, nil
}

func (s *FileService) CompleteUpload(ctx context.Context, ownerUserID, fileID int64) (*FileObject, error) {
	if s == nil || s.repo == nil || s.storage == nil {
		return nil, ErrFileStorageNotConfigured
	}
	file, err := s.repo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file.OwnerUserID != ownerUserID {
		return nil, ErrFileAccessDenied
	}
	if file.Status == FileObjectStatusPending && file.ExpiresAt != nil && !file.ExpiresAt.After(s.now()) {
		_, _ = s.repo.UpdateStatus(ctx, file.ID, FileObjectStatusExpired, nil)
		return nil, ErrFileExpired
	}
	if file.Status != FileObjectStatusPending {
		return file, nil
	}
	if err := s.storage.VerifyUploaded(ctx, file); err != nil {
		return nil, err
	}
	uploadedAt := s.now()
	return s.repo.UpdateStatus(ctx, file.ID, FileObjectStatusUploaded, &uploadedAt)
}

func isAllowedFileMIMEType(mimeType string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, item := range allowed {
		if strings.EqualFold(strings.TrimSpace(item), mimeType) {
			return true
		}
	}
	return false
}

func cleanUploadFilename(filename string) string {
	filename = strings.TrimSpace(path.Base(strings.ReplaceAll(filename, "\\", "/")))
	if filename == "" || filename == "." || filename == "/" {
		return "upload"
	}
	return filename
}

func cloneMetadata(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneFileInt64Ptr(in *int64) *int64 {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}

func cloneFileStringPtr(in *string) *string {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}
