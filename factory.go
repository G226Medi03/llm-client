package llmclient

import (
	"fmt"
	"strings"

	"github.com/G226Medi03/llm-client/deepseek"
	"github.com/G226Medi03/llm-client/ollama"
)

// Provider identifies an LLM backend.
type Provider string

const (
	ProviderDeepSeek Provider = "deepseek"
	ProviderOllama   Provider = "ollama"
)

// ParseProvider normalizes a provider name. Unknown values fall back to DeepSeek.
func ParseProvider(value string) Provider {
	switch Provider(strings.ToLower(strings.TrimSpace(value))) {
	case ProviderOllama:
		return ProviderOllama
	case ProviderDeepSeek:
		return ProviderDeepSeek
	}
	return ProviderDeepSeek
}

// Config selects a provider and configures it.
type Config struct {
	Provider Provider
	Model    string
	// APIKey is used by providers that authenticate (DeepSeek).
	APIKey string
	// BaseURL overrides the provider endpoint (Ollama host, DeepSeek API base).
	BaseURL string
	Options Options
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
