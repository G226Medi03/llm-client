package llmclient

import (
	"fmt"
	"strings"

	"github.com/G226Medi03/llm-client/internal/provider/deepseek"
	"github.com/G226Medi03/llm-client/internal/provider/gemini"
	"github.com/G226Medi03/llm-client/internal/provider/ollama"
	"github.com/G226Medi03/llm-client/internal/provider/opencode"
)

// Provider identifies an LLM backend.
type Provider string

const (
	ProviderDeepSeek Provider = "deepseek"
	ProviderOllama   Provider = "ollama"
	ProviderGemini   Provider = "gemini"
	ProviderOpenCode Provider = "opencode"
)

// Model and API constants exposed by the root facade.
const (
	ModelDeepSeekV4Flash = deepseek.ModelV4Flash
	ModelGeminiArgon     = gemini.ModelArgon
	ModelOllamaGemma4E4b = ollama.ModelGemma4E4b

	DefaultOllamaBaseURL = ollama.DefaultBaseURL

	GeminiThinkingLow    = gemini.ThinkingLow
	GeminiThinkingMedium = gemini.ThinkingMedium
	GeminiThinkingHigh   = gemini.ThinkingHigh

	OpenCodeAPIOpenAI    = opencode.APIOpenAI
	OpenCodeAPIResponses = opencode.APIResponses
	OpenCodeAPIGemini    = opencode.APIGemini
)

// ParseProvider normalizes a provider name. Unknown values fall back to DeepSeek.
func ParseProvider(value string) Provider {
	switch Provider(strings.ToLower(strings.TrimSpace(value))) {
	case ProviderOllama:
		return ProviderOllama
	case ProviderGemini:
		return ProviderGemini
	case ProviderOpenCode:
		return ProviderOpenCode
	case ProviderDeepSeek:
		return ProviderDeepSeek
	}
	return ProviderDeepSeek
}

// Config selects a provider and configures it.
type Config struct {
	Provider Provider
	Model    string
	// APIKey is used by providers that authenticate.
	APIKey string
	// BaseURL overrides the provider endpoint.
	BaseURL string
	// API is used by OpenCode: openai, responses, or gemini.
	API string
	// ThinkingLevel is used by Gemini and OpenCode's Gemini API.
	ThinkingLevel string
	// IncludeThoughts requests Gemini thought summaries.
	IncludeThoughts bool
	// MaxOutputTokens is used by Gemini.
	MaxOutputTokens int
	// OCRPrompt overrides the prompt used by DeepSeek/OpenCode OCR.
	OCRPrompt string
	// SessionID is used by OpenCode.
	SessionID string
	// UserAgent is used by OpenCode.
	UserAgent string
	Options   Options
}

// DeepSeekConfig configures a DeepSeek client.
type DeepSeekConfig struct {
	APIKey    string
	BaseURL   string
	Model     string
	OCRPrompt string
	Options   Options
}

// GeminiConfig configures a Gemini client.
type GeminiConfig struct {
	APIKey          string
	BaseURL         string
	Model           string
	ThinkingLevel   string
	IncludeThoughts bool
	MaxOutputTokens int
	Options         Options
}

// OllamaConfig configures an Ollama client.
type OllamaConfig struct {
	BaseURL string
	Model   string
	Options Options
}

// OpenCodeConfig configures an OpenCode gateway client.
type OpenCodeConfig struct {
	APIKey        string
	BaseURL       string
	Model         string
	API           string
	ThinkingLevel string
	SessionID     string
	UserAgent     string
	OCRPrompt     string
	Options       Options
}

// New builds a Client for the requested provider.
func New(cfg Config) (Client, error) {
	switch cfg.Provider {
	case ProviderOllama:
		return NewOllama(OllamaConfig{
			BaseURL: cfg.BaseURL,
			Model:   cfg.Model,
			Options: cfg.Options,
		})
	case ProviderGemini:
		return NewGemini(GeminiConfig{
			APIKey:          cfg.APIKey,
			Model:           cfg.Model,
			BaseURL:         cfg.BaseURL,
			ThinkingLevel:   cfg.ThinkingLevel,
			IncludeThoughts: cfg.IncludeThoughts,
			MaxOutputTokens: cfg.MaxOutputTokens,
			Options:         cfg.Options,
		})
	case ProviderOpenCode:
		return NewOpenCode(OpenCodeConfig{
			APIKey:        cfg.APIKey,
			BaseURL:       cfg.BaseURL,
			Model:         cfg.Model,
			API:           cfg.API,
			ThinkingLevel: cfg.ThinkingLevel,
			SessionID:     cfg.SessionID,
			UserAgent:     cfg.UserAgent,
			OCRPrompt:     cfg.OCRPrompt,
			Options:       cfg.Options,
		})
	case ProviderDeepSeek, "":
		return NewDeepSeek(DeepSeekConfig{
			APIKey:    cfg.APIKey,
			Model:     cfg.Model,
			BaseURL:   cfg.BaseURL,
			OCRPrompt: cfg.OCRPrompt,
			Options:   cfg.Options,
		})
	default:
		return nil, fmt.Errorf("llmclient: unknown provider %q", cfg.Provider)
	}
}

// NewDeepSeek builds a DeepSeek client, including OCR support.
func NewDeepSeek(cfg DeepSeekConfig) (ImageClient, error) {
	client, err := deepseek.New(deepseek.Config{
		APIKey:    cfg.APIKey,
		Model:     cfg.Model,
		BaseURL:   cfg.BaseURL,
		OCRPrompt: cfg.OCRPrompt,
		Options:   cfg.Options,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewGemini builds a Gemini client.
func NewGemini(cfg GeminiConfig) (Client, error) {
	client, err := gemini.New(gemini.Config{
		APIKey:          cfg.APIKey,
		Model:           cfg.Model,
		BaseURL:         cfg.BaseURL,
		ThinkingLevel:   cfg.ThinkingLevel,
		IncludeThoughts: cfg.IncludeThoughts,
		MaxOutputTokens: cfg.MaxOutputTokens,
		Options:         cfg.Options,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewOllama builds an Ollama client.
func NewOllama(cfg OllamaConfig) (Client, error) {
	client, err := ollama.New(ollama.Config{
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
		Options: cfg.Options,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewOpenCode builds an OpenCode gateway client, including OCR support.
func NewOpenCode(cfg OpenCodeConfig) (ImageClient, error) {
	client, err := opencode.New(opencode.Config{
		APIKey:        cfg.APIKey,
		BaseURL:       cfg.BaseURL,
		Model:         cfg.Model,
		API:           cfg.API,
		ThinkingLevel: cfg.ThinkingLevel,
		SessionID:     cfg.SessionID,
		UserAgent:     cfg.UserAgent,
		OCRPrompt:     cfg.OCRPrompt,
		Options:       cfg.Options,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
