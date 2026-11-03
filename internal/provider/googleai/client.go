// Package googleai provides a Google AI Studio generateContent client for Gemma models.
package googleai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	core "github.com/G226Medi03/llm-client/internal/core"
)

const (
	DefaultBaseURL  = "https://generativelanguage.googleapis.com/v1beta"
	DefaultModel    = "gemma-4-26b-a4b-it"
	DefaultThinking = "minimal"
)

type Config struct {
	APIKey        string
	BaseURL       string
	Model         string
	ThinkingLevel string
	HTTPClient    *http.Client
}

type Client struct {
	http          *http.Client
	baseURL       string
	model         string
	apiKey        string
	thinkingLevel string
}

type request struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type content struct {
	Role  string `json:"role"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generationConfig struct {
	Temperature     float64         `json:"temperature"`
	MaxOutputTokens int             `json:"maxOutputTokens"`
	ThinkingConfig  *thinkingConfig `json:"thinkingConfig,omitempty"`
}

type thinkingConfig struct {
	ThinkingLevel string `json:"thinkingLevel"`
}

type response struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text    string `json:"text"`
				Thought bool   `json:"thought,omitempty"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func New(cfg Config) (*Client, error) {
	key := strings.TrimSpace(cfg.APIKey)
	if key == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY 가 없습니다")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = DefaultModel
	}
	thinking := strings.TrimSpace(cfg.ThinkingLevel)
	if thinking == "" {
		thinking = DefaultThinking
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Minute}
	}
	return &Client{
		http:          httpClient,
		baseURL:       baseURL,
		model:         model,
		apiKey:        key,
		thinkingLevel: thinking,
	}, nil
}

func (c *Client) Model() string { return c.model }

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	payload, err := json.Marshal(request{
		Contents: []content{{
			Role:  "user",
			Parts: []part{{Text: prompt}},
		}},
		GenerationConfig: generationConfig{
			Temperature:     0,
			MaxOutputTokens: 128,
			ThinkingConfig:  &thinkingConfig{ThinkingLevel: c.thinkingLevel},
		},
	})
	if err != nil {
		return "", err
	}

	endpoint := c.baseURL + "/models/" + url.PathEscape(c.model) + ":generateContent"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("googleai: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("googleai: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("googleai: status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("googleai: decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("googleai: %s", parsed.Error.Message)
	}
	for _, candidate := range parsed.Candidates {
		var text strings.Builder
		for _, part := range candidate.Content.Parts {
			if part.Thought {
				continue
			}
			text.WriteString(part.Text)
		}
		if value := strings.TrimSpace(text.String()); value != "" {
			return value, nil
		}
	}
	return "", fmt.Errorf("googleai: empty response")
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
