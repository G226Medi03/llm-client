// Package llmclient is a provider-agnostic Go client for text LLMs.
//
// Providers live in subpackages (deepseek, ollama) and can also be created
// through New.
package llmclient

import "github.com/G226Medi03/llm-client/core"

type (
	// Client is a provider-agnostic text LLM client.
	Client = core.Client
	// OcrClient is implemented by clients that accept images.
	OcrClient = core.OcrClient
	// StreamResult is a single chunk emitted by Client.GenerateStream.
	StreamResult = core.StreamResult
	// Options are shared client settings.
	Options = core.Options
)
