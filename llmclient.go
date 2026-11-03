// Package llmclient is a provider-agnostic Go client for text LLMs.
//
// The root package is the only supported public API. Provider implementations
// live under internal/ and are reached through the facade constructors below.
package llmclient

import "github.com/G226Medi03/llm-client/internal/core"

type (
	// LlmClient is a provider-agnostic text LLM client.
	LlmClient = core.LlmClient
	// OcrClient is implemented by clients that accept images.
	OcrClient = core.OcrClient
	// ImageClient is an LlmClient that can also accept images.
	ImageClient interface {
		LlmClient
		OcrClient
	}
	// StreamResult is a single chunk emitted by LlmClient.GenerateStream.
	StreamResult = core.StreamResult
	// Options are shared client settings.
	Options = core.Options
)
