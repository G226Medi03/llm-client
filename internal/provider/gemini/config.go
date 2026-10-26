// Package gemini implements a Gemini API client for text LLMs.
package gemini

import (
	"errors"
	"strings"

	"github.com/G226Medi03/llm-client/internal/core"
)

// ModelArgon is the Gemini 4 thinking model.
const ModelArgon = "gemini-4-argon"

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

const (
	ThinkingLow    = "low"
	ThinkingMedium = "medium"
	ThinkingHigh   = "high"
)

// Config configures a Gemini client.
type Config struct {
	// APIKey is required.
	APIKey string
	// Model defaults to ModelArgon.
	Model string
	// BaseURL defaults to the official API.
	BaseURL string
	// ThinkingLevel is low, medium, or high. Defaults to medium.
	ThinkingLevel string
	// IncludeThoughts requests thought summaries from the API.
	IncludeThoughts bool
	// MaxOutputTokens defaults to 32768.
	MaxOutputTokens int
	// Options are shared client settings.
	Options core.Options
}

func (cfg Config) normalize() (Config, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return cfg, errors.New("gemini: APIKey is required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		cfg.Model = ModelArgon
	}
	cfg.BaseURL = strings.TrimRight(cfg.BaseURL, "/")
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	cfg.ThinkingLevel = strings.ToLower(strings.TrimSpace(cfg.ThinkingLevel))
	switch cfg.ThinkingLevel {
	case "", ThinkingMedium:
		cfg.ThinkingLevel = ThinkingMedium
	case ThinkingLow, ThinkingHigh:
	default:
		return cfg, errors.New("gemini: ThinkingLevel must be low, medium, or high")
	}
	if cfg.MaxOutputTokens <= 0 {
		cfg.MaxOutputTokens = 32768
	}
	return cfg, nil
}
