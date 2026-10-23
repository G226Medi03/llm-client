package opencode

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/G226Medi03/llm-client/core"
)

// Client is an OpenCode inference gateway client.
type Client struct {
	http          *http.Client
	baseURL       string
	apiKey        string
	model         string
	api           string
	sessionID     string
	userAgent     string
	jsonMode      bool
	ocrPrompt     string
	logger        *slog.Logger
	savedMessages []openaiMessage
}

// New builds an OpenCode client.
func New(cfg Config) (*Client, error) {
	normalized, err := cfg.normalize()
	if err != nil {
		return nil, err
	}
	return &Client{
		http:      normalized.Options.HTTPClientOrDefault(),
		baseURL:   normalized.BaseURL,
		apiKey:    normalized.APIKey,
		model:     normalized.Model,
		api:       normalized.API,
		sessionID: normalized.SessionID,
		userAgent: normalized.UserAgent,
		jsonMode:  normalized.Options.JSONMode,
		ocrPrompt: normalized.OCRPrompt,
		logger:    normalized.Options.LoggerOrDefault(),
	}, nil
}

func (c *Client) Model() string { return c.model }

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	switch c.api {
	case APIResponses:
		return c.generate_responses(ctx, prompt)
	default:
		return c.generate_openai(ctx, prompt)
	}
}

func (c *Client) GenerateStream(ctx context.Context, prompt string) (<-chan core.StreamResult, error) {
	switch c.api {
	case APIResponses:
		return c.generate_responses_stream(ctx, prompt)
	default:
		return c.generate_openai_stream(ctx, prompt)
	}
}

func (c *Client) ResetContext() { c.savedMessages = nil }

func (c *Client) Clone() core.Client { return c.clone() }

func (c *Client) clone() *Client {
	return &Client{
		http:      c.http,
		baseURL:   c.baseURL,
		apiKey:    c.apiKey,
		model:     c.model,
		api:       c.api,
		sessionID: c.sessionID,
		userAgent: c.userAgent,
		jsonMode:  c.jsonMode,
		ocrPrompt: c.ocrPrompt,
		logger:    c.logger,
	}
}

func (c *Client) headers() map[string]string {
	return map[string]string{
		"Authorization":      "Bearer " + c.apiKey,
		"x-opencode-session": c.sessionID,
		"User-Agent":         c.userAgent,
	}
}

func (c *Client) url(path string) string {
	return c.baseURL + path
}

func (c *Client) errorf(format string, args ...any) error {
	return fmt.Errorf("opencode: "+format, args...)
}
