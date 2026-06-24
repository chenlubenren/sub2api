package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestFileHandlerCreateUploadSessionReturnsUploadJSON(t *testing.T) {
	router := newFileHandlerTestRouter()
	body := `{"filename":"cat.png","mime_type":"image/png","size_bytes":12345,"purpose":"vision_input"}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/files", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var envelope struct {
		Data struct {
			FileID string `json:"file_id"`
			Status string `json:"status"`
			Upload struct {
				Method  string            `json:"method"`
				URL     string            `json:"url"`
				Headers map[string]string `json:"headers"`
			} `json:"upload"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	require.NotEmpty(t, envelope.Data.FileID)
	require.Equal(t, "pending", envelope.Data.Status)
	require.Equal(t, "PUT", envelope.Data.Upload.Method)
	require.Contains(t, envelope.Data.Upload.URL, "X-Amz-Signature")
	require.Equal(t, "image/png", envelope.Data.Upload.Headers["Content-Type"])
}

func TestFileHandlerCompleteUploadMarksUploaded(t *testing.T) {
	router := newFileHandlerTestRouter()
	fileID := createTestUploadSession(t, router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/files/"+fileID+"/complete", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var envelope struct {
		Data struct {
			FileID string `json:"file_id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	require.Equal(t, fileID, envelope.Data.FileID)
	require.Equal(t, "uploaded", envelope.Data.Status)
}

func TestFileHandlerGetReturnsFileMetadata(t *testing.T) {
	router := newFileHandlerTestRouter()
	fileID := createTestUploadSession(t, router)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/files/"+fileID, nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var envelope struct {
		Data struct {
			FileID   string `json:"file_id"`
			Status   string `json:"status"`
			MimeType string `json:"mime_type"`
			Size     int64  `json:"size_bytes"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	require.Equal(t, fileID, envelope.Data.FileID)
	require.Equal(t, "pending", envelope.Data.Status)
	require.Equal(t, "image/png", envelope.Data.MimeType)
	require.Equal(t, int64(12345), envelope.Data.Size)
}

func TestFileHandlerUnauthenticatedRequestsRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	h := NewFileHandler(newTestFileService())
	router.POST("/v1/files", h.Create)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/files", bytes.NewBufferString(`{"filename":"cat.png","mime_type":"image/png","size_bytes":12345}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func newFileHandlerTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		apiKeyID := int64(9001)
		c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{
			ID:     apiKeyID,
			UserID: 1001,
			User: &service.User{
				ID:          1001,
				Concurrency: 5,
			},
		})
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1001, Concurrency: 5})
		c.Next()
	})
	h := NewFileHandler(newTestFileService())
	router.POST("/v1/files", h.Create)
	router.POST("/v1/files/:id/complete", h.Complete)
	router.GET("/v1/files/:id", h.Get)
	return router
}

func newTestFileService() *service.FileService {
	repo := newFileHandlerFakeRepository()
	storage := &fileHandlerFakeStorage{
		upload: service.PresignedUploadInfo{
			Method: "PUT",
			URL:    "http://localhost:9000/sub2api-files/files/1001/cat.png?X-Amz-Signature=fake",
			Headers: map[string]string{
				"Content-Type": "image/png",
			},
			ExpiresAt: time.Now().Add(15 * time.Minute),
		},
	}
	return service.NewFileService(repo, storage, &config.Config{
		Storage: config.StorageConfig{
			Backend:              config.StorageBackendS3,
			Endpoint:             "http://minio:9000",
			Bucket:               "sub2api-files",
			Region:               "us-east-1",
			PresignExpireSeconds: 900,
			MaxFileSizeBytes:     10 * 1024 * 1024,
			AllowedMimeTypes:     []string{"image/png", "image/jpeg", "image/webp"},
		},
	})
}

func createTestUploadSession(t *testing.T, router *gin.Engine) string {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/files", bytes.NewBufferString(`{"filename":"cat.png","mime_type":"image/png","size_bytes":12345}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var envelope struct {
		Data struct {
			FileID string `json:"file_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &envelope))
	return envelope.Data.FileID
}

type fileHandlerFakeStorage struct {
	upload service.PresignedUploadInfo
}

func (s *fileHandlerFakeStorage) PresignUpload(_ context.Context, _ service.PresignUploadInput) (service.PresignedUploadInfo, error) {
	return s.upload, nil
}

func (s *fileHandlerFakeStorage) VerifyUploaded(_ context.Context, _ *service.FileObject) error {
	return nil
}

func (s *fileHandlerFakeStorage) PresignDownload(_ context.Context, _ *service.FileObject, _ time.Duration) (string, error) {
	return "https://files.example.com/download.png", nil
}

type fileHandlerFakeRepository struct {
	nextID int64
	files  map[int64]*service.FileObject
}

func newFileHandlerFakeRepository() *fileHandlerFakeRepository {
	return &fileHandlerFakeRepository{
		nextID: 1,
		files:  make(map[int64]*service.FileObject),
	}
}

func (r *fileHandlerFakeRepository) Create(_ context.Context, file *service.FileObject) (*service.FileObject, error) {
	cp := cloneHandlerFileObject(file)
	cp.ID = r.nextID
	r.nextID++
	now := time.Now()
	cp.CreatedAt = now
	cp.UpdatedAt = now
	r.files[cp.ID] = cp
	return cloneHandlerFileObject(cp), nil
}

func (r *fileHandlerFakeRepository) GetByID(_ context.Context, id int64) (*service.FileObject, error) {
	file, ok := r.files[id]
	if !ok {
		return nil, service.ErrFileNotFound
	}
	return cloneHandlerFileObject(file), nil
}

func (r *fileHandlerFakeRepository) UpdateStatus(_ context.Context, id int64, status string, uploadedAt *time.Time) (*service.FileObject, error) {
	file, ok := r.files[id]
	if !ok {
		return nil, service.ErrFileNotFound
	}
	file.Status = status
	file.UploadedAt = uploadedAt
	file.UpdatedAt = time.Now()
	return cloneHandlerFileObject(file), nil
}

func cloneHandlerFileObject(in *service.FileObject) *service.FileObject {
	if in == nil {
		return nil
	}
	cp := *in
	if in.APIKeyID != nil {
		v := *in.APIKeyID
		cp.APIKeyID = &v
	}
	if in.OriginalFilename != nil {
		v := *in.OriginalFilename
		cp.OriginalFilename = &v
	}
	if in.SHA256 != nil {
		v := *in.SHA256
		cp.SHA256 = &v
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

var _ service.FileRepository = (*fileHandlerFakeRepository)(nil)
var _ service.StorageProvider = (*fileHandlerFakeStorage)(nil)
