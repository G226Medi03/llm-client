package llmclient

import (
	"fmt"
	"strings"

	"github.com/G226Medi03/llm-client/internal/provider/deepseek"
	"github.com/G226Medi03/llm-client/internal/provider/gemini"
	"github.com/G226Medi03/llm-client/internal/provider/googleai"
	"github.com/G226Medi03/llm-client/internal/provider/novita"
	"github.com/G226Medi03/llm-client/internal/provider/nvidia"
	"github.com/G226Medi03/llm-client/internal/provider/ollama"
	"github.com/G226Medi03/llm-client/internal/provider/opencode"
	"github.com/G226Medi03/llm-client/internal/provider/openrouter"
)

// Provider identifies an LLM backend.
type Provider string

const (
	ProviderDeepSeek   Provider = "deepseek"
	ProviderOllama     Provider = "ollama"
	ProviderGemini     Provider = "gemini"
	ProviderOpenCode   Provider = "opencode"
	ProviderGoogleAI   Provider = "googleai"
	ProviderNovita     Provider = "novita"
	ProviderNvidia     Provider = "nvidia"
	ProviderOpenRouter Provider = "openrouter"
)

// Model and API constants exposed by the root facade.
const (
	ModelDeepSeekV4Flash = deepseek.ModelV4Flash
	ModelGeminiArgon     = gemini.ModelArgon
	ModelOllamaGemma4E4b = ollama.ModelGemma4E4b
	ModelGoogleAIGemma   = googleai.DefaultModel
	ModelNovitaGemma     = novita.DefaultModel
	ModelNvidiaGemma     = nvidia.DefaultModel
	ModelOpenRouterGemma = openrouter.DefaultModel

	DefaultGoogleAIBaseURL   = googleai.DefaultBaseURL
	DefaultNovitaBaseURL     = novita.DefaultBaseURL
	DefaultNvidiaBaseURL     = nvidia.DefaultBaseURL
	DefaultOpenRouterBaseURL = openrouter.DefaultBaseURL
	DefaultOllamaBaseURL     = ollama.DefaultBaseURL

	GoogleAIThinkingMinimal = googleai.DefaultThinking

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
	case ProviderGoogleAI:
		return ProviderGoogleAI
	case ProviderNovita:
		return ProviderNovita
	case ProviderNvidia:
		return ProviderNvidia
	case ProviderOpenRouter:
		return ProviderOpenRouter
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
	// MaxOutputTokens is used by Gemini and Ollama.
	MaxOutputTokens int
	// MaxTokens is used by Novita.
	MaxTokens int
	// Think is used by Ollama.
	Think *bool
	// HTTPReferer is used by OpenRouter.
	HTTPReferer string
	// Title is used by OpenRouter.
	Title string
	// OpenRouterProvider controls OpenRouter provider routing.
	OpenRouterProvider *OpenRouterProviderPreferences
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

// GoogleAIConfig configures a Google AI Studio client.
type GoogleAIConfig struct {
	APIKey        string
	BaseURL       string
	Model         string
	ThinkingLevel string
	Options       Options
}

// NovitaConfig configures a Novita AI client.
type NovitaConfig struct {
	APIKey    string
	BaseURL   string
	Model     string
	MaxTokens int
	Options   Options
}

// NvidiaConfig configures an NVIDIA NIM client.
type NvidiaConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Options Options
}

// OpenRouterProviderPreferences controls OpenRouter provider routing.
type OpenRouterProviderPreferences = openrouter.ProviderPreferences

// OpenRouterConfig configures an OpenRouter client.
type OpenRouterConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	HTTPReferer string
	Title       string
	Provider    *OpenRouterProviderPreferences
	Options     Options
}

// OllamaConfig configures an Ollama client.
type OllamaConfig struct {
	BaseURL         string
	Model           string
	Think           *bool
	MaxOutputTokens int
	Options         Options
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

// New builds an LlmClient for the requested provider.
func New(cfg Config) (LlmClient, error) {
	switch cfg.Provider {
	case ProviderOllama:
		return NewOllama(OllamaConfig{
			BaseURL:         cfg.BaseURL,
			Model:           cfg.Model,
			Think:           cfg.Think,
			MaxOutputTokens: cfg.MaxOutputTokens,
			Options:         cfg.Options,
		})
	case ProviderGoogleAI:
		return NewGoogleAI(GoogleAIConfig{
			APIKey:        cfg.APIKey,
			BaseURL:       cfg.BaseURL,
			Model:         cfg.Model,
			ThinkingLevel: cfg.ThinkingLevel,
			Options:       cfg.Options,
		})
	case ProviderNovita:
		return NewNovita(NovitaConfig{
			APIKey:    cfg.APIKey,
			BaseURL:   cfg.BaseURL,
			Model:     cfg.Model,
			MaxTokens: cfg.MaxTokens,
			Options:   cfg.Options,
		})
	case ProviderNvidia:
		return NewNvidia(NvidiaConfig{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL,
			Model:   cfg.Model,
			Options: cfg.Options,
		})
	case ProviderOpenRouter:
		return NewOpenRouter(OpenRouterConfig{
			APIKey:      cfg.APIKey,
			BaseURL:     cfg.BaseURL,
			Model:       cfg.Model,
			HTTPReferer: cfg.HTTPReferer,
			Title:       cfg.Title,
			Provider:    cfg.OpenRouterProvider,
			Options:     cfg.Options,
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
	return wrap_image_client(client, client), nil
}

// NewGemini builds a Gemini client.
func NewGemini(cfg GeminiConfig) (LlmClient, error) {
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
	return wrap_client(client), nil
}

// NewOllama builds an Ollama client.
func NewOllama(cfg OllamaConfig) (LlmClient, error) {
	client, err := ollama.New(ollama.Config{
		BaseURL:         cfg.BaseURL,
		Model:           cfg.Model,
		Think:           cfg.Think,
		MaxOutputTokens: cfg.MaxOutputTokens,
		Options:         cfg.Options,
	})
	if err != nil {
		return nil, err
	}
	return wrap_client(client), nil
}

// NewGoogleAI builds a Google AI Studio client.
func NewGoogleAI(cfg GoogleAIConfig) (LlmClient, error) {
	client, err := googleai.New(googleai.Config{
		APIKey:        cfg.APIKey,
		BaseURL:       cfg.BaseURL,
		Model:         cfg.Model,
		ThinkingLevel: cfg.ThinkingLevel,
		HTTPClient:    cfg.Options.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return wrap_client(client), nil
}

// NewNovita builds a Novita AI client.
func NewNovita(cfg NovitaConfig) (LlmClient, error) {
	client, err := novita.New(novita.Config{
		APIKey:     cfg.APIKey,
		BaseURL:    cfg.BaseURL,
		Model:      cfg.Model,
		MaxTokens:  cfg.MaxTokens,
		HTTPClient: cfg.Options.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return wrap_client(client), nil
}

// NewNvidia builds an NVIDIA NIM client.
func NewNvidia(cfg NvidiaConfig) (LlmClient, error) {
	client, err := nvidia.New(nvidia.Config{
		APIKey:     cfg.APIKey,
		BaseURL:    cfg.BaseURL,
		Model:      cfg.Model,
		HTTPClient: cfg.Options.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return wrap_client(client), nil
}

// NewOpenRouter builds an OpenRouter client.
func NewOpenRouter(cfg OpenRouterConfig) (LlmClient, error) {
	client, err := openrouter.New(openrouter.Config{
		APIKey:      cfg.APIKey,
		BaseURL:     cfg.BaseURL,
		Model:       cfg.Model,
		HTTPReferer: cfg.HTTPReferer,
		Title:       cfg.Title,
		Provider:    cfg.Provider,
		HTTPClient:  cfg.Options.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	return wrap_client(client), nil
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
	return wrap_image_client(client, client), nil
}
