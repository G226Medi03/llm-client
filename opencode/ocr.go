package opencode

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
)

func image_content_part(path string) (openaiContentPart, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return openaiContentPart{}, fmt.Errorf("read image: %w", err)
	}
	mime := http.DetectContentType(data)
	encoded := base64.StdEncoding.EncodeToString(data)
	return openaiContentPart{
		Type: "image_url",
		ImageURL: &openaiImageURL{
			URL: "data:" + mime + ";base64," + encoded,
		},
	}, nil
}
