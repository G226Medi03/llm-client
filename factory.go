package llmclient

import (
	"fmt"
	"strings"

	"github.com/G226Medi03/llm-client/deepseek"
	"github.com/G226Medi03/llm-client/gemini"
	"github.com/G226Medi03/llm-client/ollama"
	"github.com/G226Medi03/llm-client/opencode"
)

// Provider identifies an LLM backend.
type Provider string

const (
	ProviderDeepSeek Provider = "deepseek"
	ProviderOllama   Provider = "ollama"
	ProviderGemini   Provider = "gemini"
	ProviderOpenCode Provider = "opencode"
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
	// API is used by OpenCode: openai or responses.
	API string
	// ThinkingLevel is used by Gemini: low, medium, or high.
	ThinkingLevel string
	// IncludeThoughts requests Gemini thought summaries.
	IncludeThoughts bool
	Options         Options
}

// New builds a Client for the requested provider.
func New(cfg Config) (Client, error) {
	switch cfg.Provider {
	case ProviderOllama:
		return ollama.New(ollama.Config{
			BaseURL: cfg.BaseURL,
			Model:   cfg.Model,
			Options: cfg.Options,
		})
	case ProviderGemini:
		return gemini.New(gemini.Config{
			APIKey:          cfg.APIKey,
			Model:           cfg.Model,
			BaseURL:         cfg.BaseURL,
			ThinkingLevel:   cfg.ThinkingLevel,
			IncludeThoughts: cfg.IncludeThoughts,
			Options:         cfg.Options,
		})
	case ProviderOpenCode:
		return opencode.New(opencode.Config{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL,
			Model:   cfg.Model,
			API:     cfg.API,
			Options: cfg.Options,
		})
	case ProviderDeepSeek, "":
		return deepseek.New(deepseek.Config{
			APIKey:  cfg.APIKey,
			Model:   cfg.Model,
			BaseURL: cfg.BaseURL,
			Options: cfg.Options,
		})
	default:
		return nil, fmt.Errorf("llmclient: unknown provider %q", cfg.Provider)
	}
}
