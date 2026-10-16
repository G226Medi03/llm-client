package deepseek

import (
	"context"
	"strings"
)

// AskOCR extracts all text from an image. On failure it logs and returns "".
func (c *Client) AskOCR(ctx context.Context, imagePath string) string {
	part, err := imageContentPart(imagePath)
	if err != nil {
		c.logger.Warn("deepseek: OCR image read failed", "image", imagePath, "error", err)
		return ""
	}

	req := c.newRequest([]message{{
		Role:    "user",
		Content: []contentPart{{Type: "text", Text: c.ocrPrompt}, part},
	}}, false)

	text, err := c.send(ctx, req)
	if err != nil {
		c.logger.Warn("deepseek: OCR failed", "image", imagePath, "error", err)
		return ""
	}
	return strings.TrimSpace(text)
}

// GenerateWithImage sends a prompt together with an image.
func (c *Client) GenerateWithImage(ctx context.Context, prompt string, imagePath string) (string, error) {
	part, err := imageContentPart(imagePath)
	if err != nil {
		return "", err
	}
	req := c.newRequest([]message{{
		Role:    "user",
		Content: []contentPart{{Type: "text", Text: prompt}, part},
	}}, false)
	return c.send(ctx, req)
}
