package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/G226Medi03/llm-client/core"
	"github.com/G226Medi03/llm-client/internal/httpx"
)

// Client is a Gemini API client. It implements core.Client.
type Client struct {
	http            *http.Client
	baseURL         string
	apiKey          string
	model           string
	thinkingLevel   string
	includeThoughts bool
	maxOutputTokens int
	jsonMode        bool
	logger          *slog.Logger
}

// New builds a Gemini client.
func New(cfg Config) (*Client, error) {
	normalized, err := cfg.normalize()
	if err != nil {
		return nil, err
	}
	return &Client{
		http:            normalized.Options.HTTPClientOrDefault(),
		baseURL:         normalized.BaseURL,
		apiKey:          normalized.APIKey,
		model:           normalized.Model,
		thinkingLevel:   normalized.ThinkingLevel,
		includeThoughts: normalized.IncludeThoughts,
		maxOutputTokens: normalized.MaxOutputTokens,
		jsonMode:        normalized.Options.JSONMode,
		logger:          normalized.Options.LoggerOrDefault(),
	}, nil
}

// Model returns the configured model name.
func (c *Client) Model() string {
	return c.model
}

// Generate sends a text prompt and returns the non-thought response text.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	req := c.buildRequest(prompt)
	resp, err := httpx.PostJSON(ctx, c.http, c.generateURL(), c.headers(), req)
	if err != nil {
		return "", fmt.Errorf("gemini: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini: %s", httpx.ErrorMessage(resp))
	}

	var parsed response
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("gemini: decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("gemini: %s (%s)", parsed.Error.Message, parsed.Error.Status)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("gemini: empty response")
	}
	return text_from_parts(parsed.Candidates[0].Content.Parts), nil
}

// ResetContext is a no-op because Gemini requests are stateless.
func (c *Client) ResetContext() {}

// Clone returns a copy without mutable state.
func (c *Client) Clone() core.Client {
	return c.clone()
}

func (c *Client) clone() *Client {
	return &Client{
		http:            c.http,
		baseURL:         c.baseURL,
		apiKey:          c.apiKey,
		model:           c.model,
		thinkingLevel:   c.thinkingLevel,
		includeThoughts: c.includeThoughts,
		maxOutputTokens: c.maxOutputTokens,
		jsonMode:        c.jsonMode,
		logger:          c.logger,
	}
}

func (c *Client) generateURL() string {
	return c.baseURL + "/models/" + c.model + ":generateContent"
}

func (c *Client) streamURL() string {
	return c.baseURL + "/models/" + c.model + ":streamGenerateContent?alt=sse"
}

func (c *Client) headers() map[string]string {
	return map[string]string{"x-goog-api-key": c.apiKey}
}

func (c *Client) buildRequest(prompt string) request {
	generation_config := generationConfig{
		MaxOutputTokens: c.maxOutputTokens,
		ThinkingConfig: thinkingConfig{
			ThinkingLevel:   c.thinkingLevel,
			IncludeThoughts: c.includeThoughts,
		},
	}
	if c.jsonMode {
		generation_config.ResponseMimeType = "application/json"
	}
	return request{
		Contents: []content{{
			Role:  "user",
			Parts: []part{{Text: prompt}},
		}},
		GenerationConfig: generation_config,
	}
}

func text_from_parts(parts []part) string {
	var builder strings.Builder
	for _, item := range parts {
		if item.Thought || item.Text == "" {
			continue
		}
		builder.WriteString(item.Text)
	}
	return builder.String()
}

type request struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text    string `json:"text,omitempty"`
	Thought bool   `json:"thought,omitempty"`
}

type generationConfig struct {
	ResponseMimeType string         `json:"responseMimeType,omitempty"`
	MaxOutputTokens  int            `json:"maxOutputTokens,omitempty"`
	ThinkingConfig   thinkingConfig `json:"thinkingConfig"`
}

type thinkingConfig struct {
	ThinkingLevel   string `json:"thinkingLevel"`
	IncludeThoughts bool   `json:"includeThoughts,omitempty"`
}

type response struct {
	Candidates []candidate `json:"candidates"`
	Error      *apiError   `json:"error,omitempty"`
}

type candidate struct {
	Content      content `json:"content"`
	FinishReason string  `json:"finishReason,omitempty"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}
