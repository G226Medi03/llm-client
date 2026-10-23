// Package opencode implements an OpenCode inference gateway client.
package opencode

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/G226Medi03/llm-client/core"
)

const (
	APIOpenAI    = "openai"
	APIResponses = "responses"
	APIGemini    = "gemini"
)

const defaultBaseURL = "https://opencode.ai/inference"
const defaultModel = "deepseek-v4.1-flash"
const defaultOCRPrompt = "Extract all text from this image. Output only the text, no explanations."

// Config configures an OpenCode gateway client.
type Config struct {
	APIKey  string
	BaseURL string
	Model   string
	// API is openai, responses, or gemini.
	API string
	// ThinkingLevel is used by Gemini: low, medium, or high.
	ThinkingLevel string
	// SessionID is sent as x-opencode-session. Generated when empty.
	SessionID string
	// UserAgent identifies this client. Defaults to morbes-go-db.
	UserAgent string
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
	case APIGemini:
	default:
		return cfg, errors.New("opencode: API must be openai, responses, or gemini")
	}
	if cfg.API == APIGemini {
		cfg.ThinkingLevel = strings.ToLower(strings.TrimSpace(cfg.ThinkingLevel))
		switch cfg.ThinkingLevel {
		case "":
			cfg.ThinkingLevel = "high"
		case "low", "medium", "high":
		default:
			return cfg, errors.New("opencode: ThinkingLevel must be low, medium, or high")
		}
	}
	if strings.TrimSpace(cfg.OCRPrompt) == "" {
		cfg.OCRPrompt = defaultOCRPrompt
	}
	if strings.TrimSpace(cfg.UserAgent) == "" {
		cfg.UserAgent = "morbes-go-db/0.1.0"
	}
	if strings.TrimSpace(cfg.SessionID) == "" {
		cfg.SessionID = new_session_id()
	}
	return cfg, nil
}

func new_session_id() string {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "morbes-session"
	}
	return hex.EncodeToString(buffer)
}
