package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestFileReferenceRewriter_RewritesResponsesInputImageFileID(t *testing.T) {
	rewriter := NewFileReferenceRewriter(&fakeFileReferenceResolver{
		urls: map[int64]string{
			123: "https://files.example.com/uploaded/123.png",
		},
	})

	body := []byte(`{
		"model":"gpt-5",
		"input":[
			{
				"role":"user",
				"content":[
					{"type":"input_text","text":"describe"},
					{"type":"input_image","file_id":"123"}
				]
			}
		]
	}`)

	rewritten, changed, err := rewriter.RewriteRequestBody(context.Background(), 1001, body)

	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "https://files.example.com/uploaded/123.png", gjson.GetBytes(rewritten, "input.0.content.1.image_url").String())
	require.False(t, gjson.GetBytes(rewritten, "input.0.content.1.file_id").Exists())
}

func TestFileReferenceRewriter_RewritesChatCompletionsImageURLFileID(t *testing.T) {
	rewriter := NewFileReferenceRewriter(&fakeFileReferenceResolver{
		urls: map[int64]string{
			456: "https://files.example.com/uploaded/456.png",
		},
	})

	body := []byte(`{
		"model":"gpt-5",
		"messages":[
			{
				"role":"user",
				"content":[
					{"type":"text","text":"describe"},
					{"type":"image_url","file_id":"456","image_url":{"detail":"high"}}
				]
			}
		]
	}`)

	rewritten, changed, err := rewriter.RewriteRequestBody(context.Background(), 1001, body)

	require.NoError(t, err)
	require.True(t, changed)
	require.Equal(t, "https://files.example.com/uploaded/456.png", gjson.GetBytes(rewritten, "messages.0.content.1.image_url.url").String())
	require.Equal(t, "high", gjson.GetBytes(rewritten, "messages.0.content.1.image_url.detail").String())
	require.False(t, gjson.GetBytes(rewritten, "messages.0.content.1.file_id").Exists())
}

func TestFileReferenceRewriter_RejectsMissingFile(t *testing.T) {
	rewriter := NewFileReferenceRewriter(&fakeFileReferenceResolver{
		errs: map[int64]error{
			999: ErrFileNotFound,
		},
	})

	body := []byte(`{
		"model":"gpt-5",
		"input":[{"type":"input_image","file_id":"999"}]
	}`)

	_, _, err := rewriter.RewriteRequestBody(context.Background(), 1001, body)

	require.ErrorIs(t, err, ErrFileNotFound)
}

func TestFileReferenceRewriter_RejectsForeignOwnedFile(t *testing.T) {
	rewriter := NewFileReferenceRewriter(&fakeFileReferenceResolver{
		errs: map[int64]error{
			777: ErrFileAccessDenied,
		},
	})

	body := []byte(`{
		"model":"gpt-5",
		"messages":[
			{"role":"user","content":[{"type":"image_url","file_id":"777"}]}
		]
	}`)

	_, _, err := rewriter.RewriteRequestBody(context.Background(), 1001, body)

	require.ErrorIs(t, err, ErrFileAccessDenied)
}

type fakeFileReferenceResolver struct {
	urls map[int64]string
	errs map[int64]error
}

func (r *fakeFileReferenceResolver) ResolveInputFileURL(_ context.Context, ownerUserID, fileID int64) (string, error) {
	if ownerUserID <= 0 {
		return "", errors.New("missing owner")
	}
	if err := r.errs[fileID]; err != nil {
		return "", err
	}
	url := r.urls[fileID]
	if url == "" {
		return "", ErrFileNotFound
	}
	return url, nil
}
