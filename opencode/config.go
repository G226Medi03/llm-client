// Package opencode implements an OpenCode inference gateway client.
package opencode

import (
	"errors"
	"strings"

	"github.com/G226Medi03/llm-client/core"
)

const (
	APIOpenAI    = "openai"
	APIResponses = "responses"
)

const defaultBaseURL = "https://opencode.ai/inference"
const defaultModel = "deepseek-v4.1-flash"
const defaultOCRPrompt = "Extract all text from this image. Output only the text, no explanations."

// Config configures an OpenCode gateway client.
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
	// API is openai or responses.
	API string
	// OCRPrompt overrides the prompt used by AskOCR.
	OCRPrompt string
	Options   core.Options
}

func (cfg Config) normalize() (Config, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return cfg, errors.New("opencode: APIKey is required")
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = defaultModel
	}
	cfg.API = strings.ToLower(strings.TrimSpace(cfg.API))
	switch cfg.API {
	case "", APIOpenAI:
		cfg.API = APIOpenAI
	case APIResponses:
	default:
		return cfg, errors.New("opencode: API must be openai or responses")
	}
	if strings.TrimSpace(cfg.OCRPrompt) == "" {
		cfg.OCRPrompt = defaultOCRPrompt
	}
	return cfg, nil
}
