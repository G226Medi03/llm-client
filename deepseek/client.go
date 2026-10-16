// Package deepseek implements an OpenAI-compatible chat completions client
// for the DeepSeek API.
package deepseek

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

// ModelV4Flash is the default DeepSeek chat model.
const ModelV4Flash = "deepseek-v4-flash"

const (
	defaultBaseURL   = "https://api.deepseek.com/v1"
	defaultOCRPrompt = "Extract all text from this image. Output only the text, no explanations."
)

// Config configures a DeepSeek client.
type Config struct {
	// APIKey is required.
	APIKey string
	// Model defaults to ModelV4Flash.
	Model string
	// BaseURL defaults to the official API. The endpoint is <BaseURL>/chat/completions.
	BaseURL string
	// OCRPrompt overrides the prompt used by AskOCR.
	OCRPrompt string
	// Options are shared client settings.
	Options core.Options
}

// Client is a DeepSeek chat client. It implements core.Client and core.OcrClient.
type Client struct {
	http          *http.Client
	baseURL       string
	apiKey        string
	model         string
	jsonMode      bool
	ocrPrompt     string
	logger        *slog.Logger
	savedMessages []message
}

// New builds a DeepSeek client.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("deepseek: APIKey is required")
	}
	model := cfg.Model
	if model == "" {
		model = ModelV4Flash
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	ocrPrompt := cfg.OCRPrompt
	if ocrPrompt == "" {
		ocrPrompt = defaultOCRPrompt
	}
	return &Client{
		http:      cfg.Options.HTTPClientOrDefault(),
		baseURL:   baseURL,
		apiKey:    cfg.APIKey,
		model:     model,
		jsonMode:  cfg.Options.JSONMode,
		ocrPrompt: ocrPrompt,
		logger:    cfg.Options.LoggerOrDefault(),
	}, nil
}

// Model returns the configured model name.
func (c *Client) Model() string {
	return c.model
}

// Generate sends a text prompt and returns the response.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	req, err := c.buildRequest(prompt, "", false)
	if err != nil {
		return "", err
	}
	content, err := c.send(ctx, req)
	if err != nil {
		return "", err
	}
	c.remember(prompt, content)
	return content, nil
}

// ResetContext clears the conversation history kept for context.
func (c *Client) ResetContext() {
	c.savedMessages = nil
}

// Clone returns a copy without conversation history.
func (c *Client) Clone() core.Client {
	return c.clone()
}

func (c *Client) clone() *Client {
	return &Client{
		http:      c.http,
		baseURL:   c.baseURL,
		apiKey:    c.apiKey,
		model:     c.model,
		jsonMode:  c.jsonMode,
		ocrPrompt: c.ocrPrompt,
		logger:    c.logger,
	}
}

func (c *Client) remember(prompt, response string) {
	c.savedMessages = append(c.savedMessages,
		message{Role: "user", Content: prompt},
		message{Role: "assistant", Content: response},
	)
}

func (c *Client) buildRequest(prompt, imagePath string, stream bool) (*request, error) {
	var content any = prompt
	if imagePath != "" {
		part, err := imageContentPart(imagePath)
		if err != nil {
			return nil, fmt.Errorf("deepseek: %w", err)
		}
		content = []contentPart{{Type: "text", Text: prompt}, part}
	}

	messages := make([]message, 0, len(c.savedMessages)+1)
	messages = append(messages, c.savedMessages...)
	messages = append(messages, message{Role: "user", Content: content})

	return c.newRequest(messages, stream), nil
}

func (c *Client) newRequest(messages []message, stream bool) *request {
	req := &request{
		Model:       c.model,
		Messages:    messages,
		Stream:      stream,
		Temperature: 0,
	}
	if c.jsonMode {
		req.ResponseFormat = &responseFormat{Type: "json_object"}
	}
	return req
}

func (c *Client) chatURL() string {
	return c.baseURL + "/chat/completions"
}

func (c *Client) send(ctx context.Context, req *request) (string, error) {
	resp, err := httpx.PostJSON(ctx, c.http, c.chatURL(), map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}, req)
	if err != nil {
		return "", fmt.Errorf("deepseek: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deepseek: %s", httpx.ErrorMessage(resp))
	}

	var parsed response
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("deepseek: decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("deepseek: %s (%s)", parsed.Error.Message, parsed.Error.Type)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("deepseek: empty response")
	}
	return parsed.Choices[0].Message.Content, nil
}
