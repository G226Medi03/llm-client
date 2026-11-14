package llmclient

// LlmResponse is the text produced by one generation request.
//
// It is a distinct string type so callers can keep generated text separate
// from prompts, model names, and other application strings while still using
// it like a string.
type LlmResponse string

// String returns the response text.
func (r LlmResponse) String() string { return string(r) }
