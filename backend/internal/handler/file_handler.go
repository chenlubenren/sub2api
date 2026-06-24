package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

func (h *FileHandler) Create(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req dto.FileUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	var apiKeyID *int64
	if apiKey, ok := middleware.GetAPIKeyFromContext(c); ok && apiKey != nil {
		apiKeyID = &apiKey.ID
	}

	session, err := h.fileService.CreateUploadSession(c.Request.Context(), service.CreateUploadSessionInput{
		OwnerUserID:      subject.UserID,
		APIKeyID:         apiKeyID,
		Purpose:          req.Purpose,
		OriginalFilename: req.Filename,
		MimeType:         req.MimeType,
		SizeBytes:        req.SizeBytes,
		SHA256:           req.SHA256,
		Metadata:         req.Metadata,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, dto.FileUploadResponseFromService(session))
}

func (h *FileHandler) Complete(c *gin.Context) {
	subject, fileID, ok := h.authenticatedFileID(c)
	if !ok {
		return
	}
	file, err := h.fileService.CompleteUpload(c.Request.Context(), subject.UserID, fileID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.FileMetadataResponseFromService(file))
}

func (h *FileHandler) Get(c *gin.Context) {
	subject, fileID, ok := h.authenticatedFileID(c)
	if !ok {
		return
	}
	file, err := h.fileService.GetFile(c.Request.Context(), subject.UserID, fileID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, dto.FileMetadataResponseFromService(file))
}

func (h *FileHandler) authenticatedFileID(c *gin.Context) (middleware.AuthSubject, int64, bool) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return middleware.AuthSubject{}, 0, false
	}
	fileID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || fileID <= 0 {
		response.BadRequest(c, "Invalid file ID")
		return middleware.AuthSubject{}, 0, false
	}
	return subject, fileID, true
}
