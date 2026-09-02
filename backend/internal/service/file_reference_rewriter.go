package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type FileReferenceResolver interface {
	ResolveInputFileURL(ctx context.Context, ownerUserID, fileID int64) (string, error)
}

type FileReferenceRewriter struct {
	resolver FileReferenceResolver
}

func NewFileReferenceRewriter(resolver FileReferenceResolver) *FileReferenceRewriter {
	return &FileReferenceRewriter{resolver: resolver}
}

func (r *FileReferenceRewriter) RewriteRequestBody(ctx context.Context, ownerUserID int64, body []byte) ([]byte, bool, error) {
	if len(body) == 0 || r == nil || r.resolver == nil {
		return body, false, nil
	}

	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, false, fmt.Errorf("parse request body for file reference rewrite: %w", err)
	}

	changed, err := r.rewriteValue(ctx, ownerUserID, payload)
	if err != nil || !changed {
		return body, changed, err
	}

	rewritten, err := json.Marshal(payload)
	if err != nil {
		return nil, false, fmt.Errorf("marshal rewritten request body: %w", err)
	}
	return rewritten, true, nil
}

func (r *FileReferenceRewriter) rewriteValue(ctx context.Context, ownerUserID int64, value any) (bool, error) {
	switch typed := value.(type) {
	case map[string]any:
		changed, err := r.rewritePartIfNeeded(ctx, ownerUserID, typed)
		if err != nil {
			return false, err
		}
		for _, child := range typed {
			childChanged, childErr := r.rewriteValue(ctx, ownerUserID, child)
			if childErr != nil {
				return false, childErr
			}
			changed = changed || childChanged
		}
		return changed, nil
	case []any:
		changed := false
		for _, item := range typed {
			itemChanged, itemErr := r.rewriteValue(ctx, ownerUserID, item)
			if itemErr != nil {
				return false, itemErr
			}
			changed = changed || itemChanged
		}
		return changed, nil
	default:
		return false, nil
	}
}

func (r *FileReferenceRewriter) rewritePartIfNeeded(ctx context.Context, ownerUserID int64, part map[string]any) (bool, error) {
	switch strings.TrimSpace(firstMapString(part, "type")) {
	case "input_image":
		fileID, ok, err := parseReferencedFileID(part["file_id"])
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
		url, err := r.resolver.ResolveInputFileURL(ctx, ownerUserID, fileID)
		if err != nil {
			return false, err
		}
		part["image_url"] = url
		delete(part, "file_id")
		return true, nil
	case "image_url":
		if fileID, ok, err := parseReferencedFileID(part["file_id"]); err != nil {
			return false, err
		} else if ok {
			url, err := r.resolver.ResolveInputFileURL(ctx, ownerUserID, fileID)
			if err != nil {
				return false, err
			}
			part["image_url"] = mergeChatImageURL(part["image_url"], url)
			delete(part, "file_id")
			return true, nil
		}
		if nested, ok := part["image_url"].(map[string]any); ok {
			fileID, hasNestedFileID, err := parseReferencedFileID(nested["file_id"])
			if err != nil {
				return false, err
			}
			if hasNestedFileID {
				url, err := r.resolver.ResolveInputFileURL(ctx, ownerUserID, fileID)
				if err != nil {
					return false, err
				}
				nested["url"] = url
				delete(nested, "file_id")
				part["image_url"] = nested
				return true, nil
			}
		}
	}
	return false, nil
}

func mergeChatImageURL(current any, url string) map[string]any {
	merged, _ := current.(map[string]any)
	if merged == nil {
		merged = map[string]any{}
	}
	merged["url"] = url
	return merged
}

func parseReferencedFileID(raw any) (int64, bool, error) {
	switch value := raw.(type) {
	case nil:
		return 0, false, nil
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return 0, false, ErrFileInvalidInput
		}
		fileID, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil || fileID <= 0 {
			return 0, false, ErrFileInvalidInput
		}
		return fileID, true, nil
	case float64:
		fileID := int64(value)
		if fileID <= 0 || float64(fileID) != value {
			return 0, false, ErrFileInvalidInput
		}
		return fileID, true, nil
	default:
		return 0, false, ErrFileInvalidInput
	}
}

func firstMapString(value map[string]any, key string) string {
	raw, _ := value[key].(string)
	return raw
}
