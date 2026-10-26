// Package llmclient is a provider-agnostic Go client for text LLMs.
//
// The root package is the only supported public API. Provider implementations
// live under internal/ and are reached through the facade constructors below.
package llmclient

import "github.com/G226Medi03/llm-client/internal/core"

type (
	// Client is a provider-agnostic text LLM client.
	Client = core.Client
	// OcrClient is implemented by clients that accept images.
	OcrClient = core.OcrClient
	// ImageClient is a Client that can also accept images.
	ImageClient interface {
		Client
		OcrClient
	}
	// StreamResult is a single chunk emitted by Client.GenerateStream.
	StreamResult = core.StreamResult
	// Options are shared client settings.
	Options = core.Options
)
