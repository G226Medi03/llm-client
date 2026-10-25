// Package call provides isolated single-call execution for llmclient clients.
package call

import (
	"context"
	"fmt"
	"time"

	llmclient "github.com/G226Medi03/llm-client"
)

// Generate runs one LLM request with an isolated client context and deadline.
// The base client is never mutated, so parallel callers cannot share conversation state.
func Generate(ctx context.Context, base llmclient.Client, prompt string, timeout time.Duration) (string, error) {
	if base == nil {
		return "", fmt.Errorf("LLM client 가 없습니다")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	client := base.Clone()
	client.ResetContext()
	defer client.ResetContext()

	if timeout <= 0 {
		return client.Generate(ctx, prompt)
	}

	call_ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return client.Generate(call_ctx, prompt)
}
