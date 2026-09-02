package handler

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestRewriteAndValidateOpenAIRequestBody_RewritesFileIDBeforeInlineValidation(t *testing.T) {
	rewriter := service.NewFileReferenceRewriter(&requestBodyPreprocessResolver{
		urls: map[int64]string{
			123: "https://files.example.com/123.png",
		},
	})

	body := []byte(`{
		"model":"gpt-5",
		"input":[
			{"type":"input_image","file_id":"123"}
		]
	}`)

	rewritten, err := rewriteAndValidateOpenAIRequestBody(context.Background(), rewriter, 1, 1001, body)

	require.NoError(t, err)
	require.Equal(t, "https://files.example.com/123.png", gjson.GetBytes(rewritten, "input.0.image_url").String())
	require.False(t, gjson.GetBytes(rewritten, "input.0.file_id").Exists())
}

func TestRewriteAndValidateOpenAIRequestBody_RejectsOversizedInlineImage(t *testing.T) {
	body := []byte(`{
		"model":"gpt-5",
		"input":[
			{"type":"input_image","image_url":"data:image/png;base64,QUJDREVGR0g="}
		]
	}`)

	_, err := rewriteAndValidateOpenAIRequestBody(context.Background(), nil, 4, 1001, body)

	require.Error(t, err)
	require.Contains(t, err.Error(), "/v1/files")
}

type requestBodyPreprocessResolver struct {
	urls map[int64]string
}

func (r *requestBodyPreprocessResolver) ResolveInputFileURL(_ context.Context, _ int64, fileID int64) (string, error) {
	return r.urls[fileID], nil
}
