package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

func rewriteAndValidateOpenAIRequestBody(
	ctx context.Context,
	rewriter *service.FileReferenceRewriter,
	maxInlineImageBytes int64,
	ownerUserID int64,
	body []byte,
) ([]byte, error) {
	rewritten := body
	if rewriter != nil {
		var err error
		rewritten, _, err = rewriter.RewriteRequestBody(ctx, ownerUserID, body)
		if err != nil {
			return nil, err
		}
	}
	if err := service.ValidateInlineImageDataURIsForCompatibility(rewritten, maxInlineImageBytes); err != nil {
		return nil, err
	}
	return rewritten, nil
}

func resolveMaxInlineImageBytes(cfg *config.Config) int64 {
	if cfg != nil && cfg.Gateway.MaxInlineImageBytes > 0 {
		return cfg.Gateway.MaxInlineImageBytes
	}
	return config.DefaultGatewayMaxInlineImageBytes
}

func requestBodyPreprocessErrorDetails(err error) (status int, code, errType, message string) {
	status = infraerrors.Code(err)
	if status <= 0 {
		status = http.StatusBadRequest
	}
	code = strings.TrimSpace(infraerrors.Reason(err))
	if code == "" {
		code = "invalid_request_error"
	}
	message = strings.TrimSpace(infraerrors.Message(err))
	if message == "" {
		message = "Failed to preprocess request body"
	}
	errType = "invalid_request_error"
	switch status {
	case http.StatusUnauthorized:
		errType = "authentication_error"
	case http.StatusForbidden:
		errType = "permission_error"
	default:
		if status >= 500 {
			errType = "api_error"
		}
	}
	return status, code, errType, message
}
