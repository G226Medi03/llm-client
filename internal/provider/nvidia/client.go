// Package nvidia provides an OpenAI-compatible NVIDIA NIM text client.
package nvidia

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
	DefaultBaseURL = "https://integrate.api.nvidia.com/v1"
	DefaultModel   = "google/gemma-4-31b-it"
)

type Config struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

type Client struct {
	http    *http.Client
	baseURL string
	apiKey  string
	model   string
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
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func New(cfg Config) (*Client, error) {
	key := strings.TrimSpace(cfg.APIKey)
	if key == "" {
		return nil, fmt.Errorf("NVIDIA_API_KEY 가 없습니다")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = DefaultModel
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Minute}
	}
	return &Client{
		http:    httpClient,
		baseURL: baseURL,
		apiKey:  key,
		model:   model,
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
		MaxTokens:   64,
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
		return "", fmt.Errorf("nvidia: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("nvidia: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nvidia: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("nvidia: decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("nvidia: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("nvidia: empty response")
	}
	return parsed.Choices[0].Message.Content, nil
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

var _ core.ProviderClient = (*Client)(nil)
