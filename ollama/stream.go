package ollama

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/G226Medi03/llmclient/core"
	"github.com/G226Medi03/llmclient/internal/httpx"
)

// GenerateStream sends a text prompt and streams the response chunks.
func (c *Client) GenerateStream(ctx context.Context, prompt string) (<-chan core.StreamResult, error) {
	body := c.newRequest(prompt, true)
	resp, err := httpx.PostJSON(ctx, c.http, c.chatURL(), nil, body)
	if err != nil {
		return nil, fmt.Errorf("ollama: %w", err)
	}

	out := make(chan core.StreamResult)

	go func() {
		defer resp.Body.Close()
		defer close(out)

		if resp.StatusCode != http.StatusOK {
			out <- core.StreamResult{Err: fmt.Errorf("ollama: %s", httpx.ErrorMessage(resp))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var chunk chatResponse
			if err := json.Unmarshal([]byte(line), &chunk); err != nil {
				out <- core.StreamResult{Err: fmt.Errorf("ollama: decode stream chunk: %w", err)}
				return
			}
			if chunk.Error != "" {
				out <- core.StreamResult{Err: fmt.Errorf("ollama: %s", chunk.Error)}
				return
			}
			if chunk.Message.Content != "" {
				out <- core.StreamResult{Response: chunk.Message.Content}
			}
			if chunk.Done {
				break
			}
		}
		if err := scanner.Err(); err != nil {
			out <- core.StreamResult{Err: err}
		}
	}()

	return out, nil
}
