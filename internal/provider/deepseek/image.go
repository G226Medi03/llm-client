package deepseek

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func imageMimeType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg"
	}
}

func imageContentPart(imagePath string) (contentPart, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return contentPart{}, fmt.Errorf("read image %s: %w", imagePath, err)
	}
	dataURL := fmt.Sprintf("data:%s;base64,%s", imageMimeType(imagePath), base64.StdEncoding.EncodeToString(data))
	return contentPart{
		Type:     "image_url",
		ImageURL: &imageURL{URL: dataURL, Detail: "high"},
	}, nil
}
