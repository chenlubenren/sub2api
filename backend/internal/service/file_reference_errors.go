package service

import infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"

var (
	ErrFileNotFound             = infraerrors.NotFound("FILE_NOT_FOUND", "file not found")
	ErrFileAccessDenied         = infraerrors.Forbidden("FILE_ACCESS_DENIED", "file access denied")
	ErrFileExpired              = infraerrors.BadRequest("FILE_EXPIRED", "file upload session expired")
	ErrFileInvalidInput         = infraerrors.BadRequest("FILE_INVALID_INPUT", "invalid file input")
	ErrFileStorageNotConfigured = infraerrors.BadRequest("FILE_STORAGE_NOT_CONFIGURED", "file storage is not configured")
)
