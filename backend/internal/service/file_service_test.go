package service

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestFileServiceCreateUploadSessionReturnsFileIDAndPresignedUploadInfo(t *testing.T) {
	ctx := context.Background()
	repo := newFakeFileRepository()
	storage := &fakeStorageProvider{
		upload: PresignedUploadInfo{
			Method: "PUT",
			URL:    "http://public.example.com/sub2api-files/files/1/test.png?X-Amz-Signature=fake",
			Headers: map[string]string{
				"Content-Type": "image/png",
			},
			ExpiresAt: time.Now().Add(15 * time.Minute),
		},
	}
	svc := NewFileService(repo, storage, testFileStorageConfig())

	session, err := svc.CreateUploadSession(ctx, CreateUploadSessionInput{
		OwnerUserID:      1001,
		APIKeyID:         fileInt64Ptr(9001),
		Purpose:          "vision_input",
		OriginalFilename: "test.png",
		MimeType:         "image/png",
		SizeBytes:        12345,
		Metadata:         map[string]any{"source": "unit-test"},
	})

	require.NoError(t, err)
	require.NotZero(t, session.FileID)
	require.Equal(t, "PUT", session.Upload.Method)
	require.Contains(t, session.Upload.URL, "X-Amz-Signature")
	require.Equal(t, "image/png", session.Upload.Headers["Content-Type"])
	require.Equal(t, FileObjectStatusPending, repo.files[session.FileID].Status)
	require.Equal(t, "sub2api-files", repo.files[session.FileID].Bucket)
	require.NotEmpty(t, repo.files[session.FileID].ObjectKey)
	require.Equal(t, repo.files[session.FileID].ObjectKey, storage.lastObjectKey)
}

func TestFileServiceCompleteUploadTransitionsStatusFromPendingToUploaded(t *testing.T) {
	ctx := context.Background()
	repo := newFakeFileRepository()
	storage := &fakeStorageProvider{upload: fakeUploadInfo()}
	svc := NewFileService(repo, storage, testFileStorageConfig())
	session, err := svc.CreateUploadSession(ctx, CreateUploadSessionInput{
		OwnerUserID:      1001,
		OriginalFilename: "test.png",
		MimeType:         "image/png",
		SizeBytes:        12345,
	})
	require.NoError(t, err)

	file, err := svc.CompleteUpload(ctx, 1001, session.FileID)

	require.NoError(t, err)
	require.Equal(t, FileObjectStatusUploaded, file.Status)
	require.NotNil(t, file.UploadedAt)
	require.Equal(t, FileObjectStatusUploaded, repo.files[session.FileID].Status)
	require.True(t, storage.verifyCalled)
}

func TestFileServiceUnauthorizedFileAccessRejected(t *testing.T) {
	ctx := context.Background()
	repo := newFakeFileRepository()
	storage := &fakeStorageProvider{upload: fakeUploadInfo()}
	svc := NewFileService(repo, storage, testFileStorageConfig())
	session, err := svc.CreateUploadSession(ctx, CreateUploadSessionInput{
		OwnerUserID:      1001,
		OriginalFilename: "test.png",
		MimeType:         "image/png",
		SizeBytes:        12345,
	})
	require.NoError(t, err)

	_, err = svc.CompleteUpload(ctx, 2002, session.FileID)

	require.ErrorIs(t, err, ErrFileAccessDenied)
	require.False(t, storage.verifyCalled)
	require.Equal(t, FileObjectStatusPending, repo.files[session.FileID].Status)
}

func TestFileServiceExpiredPendingFileRejected(t *testing.T) {
	ctx := context.Background()
	repo := newFakeFileRepository()
	storage := &fakeStorageProvider{upload: fakeUploadInfo()}
	svc := NewFileService(repo, storage, testFileStorageConfig())
	expiredAt := time.Now().Add(-time.Minute)
	file, err := repo.Create(ctx, &FileObject{
		OwnerUserID:      1001,
		Purpose:          "vision_input",
		StorageProvider:  "s3",
		Bucket:           "sub2api-files",
		ObjectKey:        "files/1001/expired.png",
		OriginalFilename: fileStringPtr("expired.png"),
		MimeType:         "image/png",
		SizeBytes:        12345,
		Status:           FileObjectStatusPending,
		ExpiresAt:        &expiredAt,
	})
	require.NoError(t, err)

	_, err = svc.CompleteUpload(ctx, 1001, file.ID)

	require.ErrorIs(t, err, ErrFileExpired)
	require.False(t, storage.verifyCalled)
	require.Equal(t, FileObjectStatusExpired, repo.files[file.ID].Status)
}

func TestFileServiceS3PresignUsesEndpointWhenPublicEndpointEmpty(t *testing.T) {
	ctx := context.Background()
	cfg := testFileStorageConfig()
	cfg.Storage.Endpoint = "http://minio:9000"
	cfg.Storage.PublicEndpoint = ""
	provider, err := NewS3StorageProvider(ctx, cfg)
	require.NoError(t, err)

	upload, err := provider.PresignUpload(ctx, PresignUploadInput{
		Bucket:      "sub2api-files",
		ObjectKey:   "files/1001/test.png",
		ContentType: "image/png",
		Expires:     15 * time.Minute,
	})

	require.NoError(t, err)
	parsed, err := url.Parse(upload.URL)
	require.NoError(t, err)
	require.Equal(t, "minio:9000", parsed.Host)
}

func TestFileServiceS3PresignUsesPublicEndpointWhenConfigured(t *testing.T) {
	ctx := context.Background()
	cfg := testFileStorageConfig()
	cfg.Storage.Endpoint = "http://minio:9000"
	cfg.Storage.PublicEndpoint = "http://localhost:9000"
	provider, err := NewS3StorageProvider(ctx, cfg)
	require.NoError(t, err)

	upload, err := provider.PresignUpload(ctx, PresignUploadInput{
		Bucket:      "sub2api-files",
		ObjectKey:   "files/1001/test.png",
		ContentType: "image/png",
		Expires:     15 * time.Minute,
	})

	require.NoError(t, err)
	parsed, err := url.Parse(upload.URL)
	require.NoError(t, err)
	require.Equal(t, "localhost:9000", parsed.Host)
}

type fakeFileRepository struct {
	nextID int64
	files  map[int64]*FileObject
}

func newFakeFileRepository() *fakeFileRepository {
	return &fakeFileRepository{
		nextID: 1,
		files:  make(map[int64]*FileObject),
	}
}

func (r *fakeFileRepository) Create(_ context.Context, file *FileObject) (*FileObject, error) {
	cp := cloneFileObject(file)
	cp.ID = r.nextID
	r.nextID++
	now := time.Now()
	cp.CreatedAt = now
	cp.UpdatedAt = now
	r.files[cp.ID] = cp
	return cloneFileObject(cp), nil
}

func (r *fakeFileRepository) GetByID(_ context.Context, id int64) (*FileObject, error) {
	file, ok := r.files[id]
	if !ok {
		return nil, ErrFileNotFound
	}
	return cloneFileObject(file), nil
}

func (r *fakeFileRepository) UpdateStatus(_ context.Context, id int64, status string, uploadedAt *time.Time) (*FileObject, error) {
	file, ok := r.files[id]
	if !ok {
		return nil, ErrFileNotFound
	}
	file.Status = status
	file.UploadedAt = uploadedAt
	file.UpdatedAt = time.Now()
	return cloneFileObject(file), nil
}

type fakeStorageProvider struct {
	upload        PresignedUploadInfo
	verifyErr     error
	verifyCalled  bool
	lastObjectKey string
}

func (p *fakeStorageProvider) PresignUpload(_ context.Context, input PresignUploadInput) (PresignedUploadInfo, error) {
	p.lastObjectKey = input.ObjectKey
	if p.upload.Headers == nil {
		p.upload.Headers = map[string]string{"Content-Type": input.ContentType}
	}
	return p.upload, nil
}

func (p *fakeStorageProvider) VerifyUploaded(_ context.Context, _ *FileObject) error {
	p.verifyCalled = true
	if p.verifyErr != nil {
		return p.verifyErr
	}
	return nil
}

func testFileStorageConfig() *config.Config {
	return &config.Config{
		Storage: config.StorageConfig{
			Backend:              config.StorageBackendS3,
			Endpoint:             "http://minio:9000",
			Region:               "us-east-1",
			Bucket:               "sub2api-files",
			AccessKey:            "minioadmin",
			SecretKey:            "minioadmin",
			UsePathStyle:         true,
			PresignExpireSeconds: 900,
			MaxFileSizeBytes:     10 * 1024 * 1024,
			AllowedMimeTypes:     []string{"image/png", "image/jpeg", "image/webp"},
		},
	}
}

func fakeUploadInfo() PresignedUploadInfo {
	return PresignedUploadInfo{
		Method:    "PUT",
		URL:       "http://public.example.com/sub2api-files/files/1/test.png?X-Amz-Signature=fake",
		Headers:   map[string]string{"Content-Type": "image/png"},
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
}

func cloneFileObject(in *FileObject) *FileObject {
	if in == nil {
		return nil
	}
	cp := *in
	if in.APIKeyID != nil {
		cp.APIKeyID = fileInt64Ptr(*in.APIKeyID)
	}
	if in.OriginalFilename != nil {
		cp.OriginalFilename = fileStringPtr(*in.OriginalFilename)
	}
	if in.SHA256 != nil {
		cp.SHA256 = fileStringPtr(*in.SHA256)
	}
	if in.UploadedAt != nil {
		v := *in.UploadedAt
		cp.UploadedAt = &v
	}
	if in.ExpiresAt != nil {
		v := *in.ExpiresAt
		cp.ExpiresAt = &v
	}
	if in.Metadata != nil {
		cp.Metadata = make(map[string]any, len(in.Metadata))
		for k, v := range in.Metadata {
			cp.Metadata[k] = v
		}
	}
	return &cp
}

func fileInt64Ptr(v int64) *int64 {
	return &v
}

func fileStringPtr(v string) *string {
	return &v
}

var _ FileRepository = (*fakeFileRepository)(nil)
var _ StorageProvider = (*fakeStorageProvider)(nil)

func TestFileServiceStorageVerifierErrorIsReturned(t *testing.T) {
	ctx := context.Background()
	repo := newFakeFileRepository()
	verifyErr := errors.New("head object failed")
	storage := &fakeStorageProvider{upload: fakeUploadInfo(), verifyErr: verifyErr}
	svc := NewFileService(repo, storage, testFileStorageConfig())
	session, err := svc.CreateUploadSession(ctx, CreateUploadSessionInput{
		OwnerUserID:      1001,
		OriginalFilename: "test.png",
		MimeType:         "image/png",
		SizeBytes:        12345,
	})
	require.NoError(t, err)

	_, err = svc.CompleteUpload(ctx, 1001, session.FileID)

	require.ErrorIs(t, err, verifyErr)
	require.Equal(t, FileObjectStatusPending, repo.files[session.FileID].Status)
}
