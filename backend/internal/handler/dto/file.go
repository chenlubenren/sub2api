package dto

import (
	"strconv"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

type FileUploadRequest struct {
	Filename  string         `json:"filename" binding:"required"`
	MimeType  string         `json:"mime_type" binding:"required"`
	SizeBytes int64          `json:"size_bytes" binding:"required,min=0"`
	Purpose   string         `json:"purpose"`
	SHA256    *string        `json:"sha256"`
	Metadata  map[string]any `json:"metadata"`
}

type FileUploadResponse struct {
	FileID    string            `json:"file_id"`
	Status    string            `json:"status"`
	MimeType  string            `json:"mime_type"`
	SizeBytes int64             `json:"size_bytes"`
	ExpiresAt *time.Time        `json:"expires_at,omitempty"`
	Upload    FileUploadInfoDTO `json:"upload"`
}

type FileUploadInfoDTO struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt time.Time         `json:"expires_at"`
}

type FileMetadataResponse struct {
	FileID    string     `json:"file_id"`
	Status    string     `json:"status"`
	MimeType  string     `json:"mime_type"`
	SizeBytes int64      `json:"size_bytes"`
	Purpose   string     `json:"purpose"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func FileUploadResponseFromService(session *service.UploadSession) FileUploadResponse {
	if session == nil || session.File == nil {
		return FileUploadResponse{}
	}
	return FileUploadResponse{
		FileID:    strconv.FormatInt(session.FileID, 10),
		Status:    session.File.Status,
		MimeType:  session.File.MimeType,
		SizeBytes: session.File.SizeBytes,
		ExpiresAt: session.File.ExpiresAt,
		Upload: FileUploadInfoDTO{
			Method:    session.Upload.Method,
			URL:       session.Upload.URL,
			Headers:   session.Upload.Headers,
			ExpiresAt: session.Upload.ExpiresAt,
		},
	}
}

func FileMetadataResponseFromService(file *service.FileObject) FileMetadataResponse {
	if file == nil {
		return FileMetadataResponse{}
	}
	return FileMetadataResponse{
		FileID:    strconv.FormatInt(file.ID, 10),
		Status:    file.Status,
		MimeType:  file.MimeType,
		SizeBytes: file.SizeBytes,
		Purpose:   file.Purpose,
		CreatedAt: file.CreatedAt,
		ExpiresAt: file.ExpiresAt,
	}
}
