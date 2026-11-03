// Package novita provides an OpenAI-compatible Novita AI text client.
package novita

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	core "github.com/G226Medi03/llm-client/internal/core"
)

const (
	DefaultBaseURL   = "https://api.novita.ai/openai/v1"
	DefaultModel     = "google/gemma-4-31b-it"
	DefaultMaxTokens = 128
)

type Config struct {
	APIKey     string
	BaseURL    string
	Model      string
	MaxTokens  int
	HTTPClient *http.Client
}

type Client struct {
	http      *http.Client
	baseURL   string
	apiKey    string
	model     string
	maxTokens int
}

type request struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type response struct {
	Choices []struct {
		Message struct {
			Content   string `json:"content"`
			Reasoning string `json:"reasoning_content,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

func New(cfg Config) (*Client, error) {
	key := strings.TrimSpace(cfg.APIKey)
	if key == "" {
		return nil, fmt.Errorf("NOVITA_API_KEY 가 없습니다")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = DefaultModel
	}
	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Minute}
	}
	return &Client{
		http:      httpClient,
		baseURL:   baseURL,
		apiKey:    key,
		model:     model,
		maxTokens: maxTokens,
	}, nil
}

func (c *Client) Model() string { return c.model }

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	payload, err := json.Marshal(request{
		Model: c.model,
		Messages: []message{{
			Role:    "user",
			Content: prompt,
		}},
		Temperature: 0,
		MaxTokens:   c.maxTokens,
		Stream:      false,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("novita: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("novita: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("novita: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("novita: decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("novita: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("novita: empty response")
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("novita: empty content (reasoning=%q)", truncate(parsed.Choices[0].Message.Reasoning, 120))
	}
	return content, nil
}

func (c *Client) GenerateStream(ctx context.Context, prompt string) (<-chan core.StreamResult, error) {
	out := make(chan core.StreamResult, 1)
	go func() {
		defer close(out)
		text, err := c.Generate(ctx, prompt)
		if err != nil {
			out <- core.StreamResult{Err: err}
			return
		}
		out <- core.StreamResult{Response: text}
	}()
	return out, nil
}

func (c *Client) ResetContext() {}

func (c *Client) Clone() core.ProviderClient {
	clone := *c
	return &clone
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "…"
}

var _ core.ProviderClient = (*Client)(nil)
