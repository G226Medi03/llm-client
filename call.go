package llmclient

import (
	"context"
	"fmt"
	"time"
)

// GenerateIsolated runs one LLM request with an isolated client context and
// deadline. The base client is never mutated, so parallel callers cannot share
// conversation state.
func GenerateIsolated(ctx context.Context, base Client, prompt string, timeout time.Duration) (string, error) {
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
