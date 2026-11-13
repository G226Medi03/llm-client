package call

import (
	"context"

	llmclient "github.com/G226Medi03/llm-client"
)

type GenerateFunc func(ctx context.Context, prompt string) (string, error)

// Client is the application-level text generation unit. It is not a provider
// port: it owns replay, retry, timeout, and test behavior while delegating
// actual network calls through FromLLMClient.
type Client struct {
	model    string
	generate GenerateFunc
	clone    func() *Client
	reset    func()
}

func NewFunc(model string, generate GenerateFunc) *Client {
	if generate == nil {
		return nil
	}
	client := &Client{model: model, generate: generate, reset: func() {}}
	client.clone = func() *Client {
		return NewFunc(client.model, client.generate)
	}
	return client
}

func FromLLMClient(client llmclient.LlmClient) *Client {
	if client == nil {
		return nil
	}
	result := &Client{
		model: client.Model(),
		generate: func(ctx context.Context, prompt string) (string, error) {
			return llmclient.GenerateIsolated(ctx, client, prompt, LLM_REQUEST_TIMEOUT())
		},
		reset: client.ResetContext,
	}
	result.clone = func() *Client {
		return FromLLMClient(client.Clone())
	}
	return result
}

func (c *Client) Model() string {
	if c == nil {
		return ""
	}
	return c.model
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	if c == nil || c.generate == nil {
		return "", context.Canceled
	}
	return generateWithPolicy(ctx, c.generate, prompt)
}

func (c *Client) Clone() *Client {
	if c == nil {
		return nil
	}
	if c.clone != nil {
		return c.clone()
	}
	return NewFunc(c.model, c.generate)
}

func (c *Client) ResetContext() {
	if c != nil && c.reset != nil {
		c.reset()
	}
}
