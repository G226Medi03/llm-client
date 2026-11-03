// Package llmclient is a provider-agnostic Go client for text LLMs.
//
// The root package is the only supported public API. Provider implementations
// live under internal/ and are reached through the facade constructors below.
package llmclient

import (
	"context"

	"github.com/G226Medi03/llm-client/internal/core"
)

// LlmClient is a provider-agnostic text LLM client.
// The unexported method seals the interface so only this module can provide
// implementations.
type LlmClient interface {
	Model() string
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateStream(ctx context.Context, prompt string) (<-chan StreamResult, error)
	ResetContext()
	Clone() LlmClient

	is_LlmClient()
}

// OcrClient is implemented by clients that accept images.
type OcrClient = core.OcrClient

// ImageClient is an LlmClient that can also accept images.
type ImageClient interface {
	LlmClient
	OcrClient
}

// StreamResult is a single chunk emitted by LlmClient.GenerateStream.
type StreamResult = core.StreamResult

// Options are shared client settings.
type Options = core.Options

type clientWrapper struct {
	inner core.ProviderClient
}

func (c *clientWrapper) is_LlmClient() {}

func (c *clientWrapper) Model() string {
	return c.inner.Model()
}

func (c *clientWrapper) Generate(ctx context.Context, prompt string) (string, error) {
	return c.inner.Generate(ctx, prompt)
}

func (c *clientWrapper) GenerateStream(ctx context.Context, prompt string) (<-chan StreamResult, error) {
	return c.inner.GenerateStream(ctx, prompt)
}

func (c *clientWrapper) ResetContext() {
	c.inner.ResetContext()
}

func (c *clientWrapper) Clone() LlmClient {
	return &clientWrapper{inner: c.inner.Clone()}
}

type imageClientWrapper struct {
	*clientWrapper
	ocr core.OcrClient
}

func (c *imageClientWrapper) AskOCR(ctx context.Context, imagePath string) string {
	return c.ocr.AskOCR(ctx, imagePath)
}

func (c *imageClientWrapper) GenerateWithImage(
	ctx context.Context,
	prompt string,
	imagePath string,
) (string, error) {
	return c.ocr.GenerateWithImage(ctx, prompt, imagePath)
}

func (c *imageClientWrapper) Clone() LlmClient {
	cloned := c.inner.Clone()
	if ocr, ok := cloned.(core.OcrClient); ok {
		return &imageClientWrapper{
			clientWrapper: &clientWrapper{inner: cloned},
			ocr:           ocr,
		}
	}
	return &clientWrapper{inner: cloned}
}

func wrap_client(inner core.ProviderClient) LlmClient {
	return &clientWrapper{inner: inner}
}

func wrap_image_client(inner core.ProviderClient, ocr core.OcrClient) ImageClient {
	return &imageClientWrapper{
		clientWrapper: &clientWrapper{inner: inner},
		ocr:           ocr,
	}
}
