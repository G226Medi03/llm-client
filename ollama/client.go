// Package ollama implements a client for the Ollama /api/chat endpoint.
package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/G226Medi03/llm-client/core"
	"github.com/G226Medi03/llm-client/internal/httpx"
)

const (
	// DefaultBaseURL is the default local Ollama endpoint.
	DefaultBaseURL = "http://localhost:11434"
	// ModelGemma4E4b is a common small local model.
	ModelGemma4E4b = "gemma4:e4b"
)

// Config configures an Ollama client.
type Config struct {
	// BaseURL defaults to DefaultBaseURL.
	BaseURL string
	// Model is required.
	Model string
	// Options are shared client settings.
	Options core.Options
}

// Client is an Ollama chat client. It implements core.Client.
type Client struct {
	http     *http.Client
	baseURL  string
	model    string
	jsonMode bool
	logger   *slog.Logger
}

// New builds an Ollama client.
func New(cfg Config) (*Client, error) {
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		return nil, fmt.Errorf("ollama: Model is required")
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		http:     cfg.Options.HTTPClientOrDefault(),
		baseURL:  baseURL,
		model:    model,
		jsonMode: cfg.Options.JSONMode,
		logger:   cfg.Options.LoggerOrDefault(),
	}, nil
}

// Model returns the configured model name.
func (c *Client) Model() string {
	return c.model
}

// Generate sends a text prompt and returns the response.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	body := c.newRequest(prompt, false)
	resp, err := httpx.PostJSON(ctx, c.http, c.chatURL(), nil, body)
	if err != nil {
		return "", fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: %s", httpx.ErrorMessage(resp))
	}

	var parsed chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("ollama: decode response: %w", err)
	}
	if parsed.Error != "" {
		return "", fmt.Errorf("ollama: %s", parsed.Error)
	}
	return parsed.Message.Content, nil
}

// ResetContext is a no-op for Ollama (it is stateless by request).
func (c *Client) ResetContext() {}

// Clone returns an equivalent client.
func (c *Client) Clone() core.Client {
	return &Client{
		http:     c.http,
		baseURL:  c.baseURL,
		model:    c.model,
		jsonMode: c.jsonMode,
		logger:   c.logger,
	}
}

func (c *Client) chatURL() string {
	return c.baseURL + "/api/chat"
}

func (c *Client) newRequest(prompt string, stream bool) chatRequest {
	req := chatRequest{
		Model:    c.model,
		Messages: []chatMessage{{Role: "user", Content: prompt}},
		Stream:   stream,
	}
	if c.jsonMode {
		req.Format = "json"
	}
	return req
}
