// Package core defines the provider-agnostic types shared by llmclient providers.
package core

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// StreamResult is a single chunk emitted by Client.GenerateStream.
type StreamResult struct {
	Response string
	Err      error
}

// Client is a provider-agnostic text LLM client.
type Client interface {
	Model() string
	Generate(ctx context.Context, prompt string) (string, error)
	GenerateStream(ctx context.Context, prompt string) (<-chan StreamResult, error)
	ResetContext()
	Clone() Client
}

// OcrClient is implemented by clients that accept images.
type OcrClient interface {
	// AskOCR extracts all text from the image. On failure it returns "".
	AskOCR(ctx context.Context, imagePath string) string
	// GenerateWithImage sends a prompt together with an image.
	GenerateWithImage(ctx context.Context, prompt string, imagePath string) (string, error)
}

// Options are shared client settings. The zero value is usable.
type Options struct {
	// JSONMode asks the provider to return JSON only.
	JSONMode bool
	// Logger receives warnings. Nil means no logging.
	Logger *slog.Logger
	// HTTPClient overrides the default HTTP client.
	HTTPClient *http.Client
}

// LoggerOrDefault returns the configured logger or a discarding one.
func (o Options) LoggerOrDefault() *slog.Logger {
	if o.Logger != nil {
		return o.Logger
	}
	return slog.New(slog.DiscardHandler)
}

// HTTPClientOrDefault returns the configured HTTP client or a default one.
func (o Options) HTTPClientOrDefault() *http.Client {
	if o.HTTPClient != nil {
		return o.HTTPClient
	}
	return &http.Client{Timeout: 3 * time.Minute}
}
