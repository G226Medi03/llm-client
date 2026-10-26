package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/G226Medi03/llm-client/internal/core"
	"github.com/G226Medi03/llm-client/internal/httpx"
)

// GenerateStream sends a text prompt and streams non-thought response text.
func (c *Client) GenerateStream(ctx context.Context, prompt string) (<-chan core.StreamResult, error) {
	req := c.buildRequest(prompt)
	resp, err := httpx.PostJSON(ctx, c.http, c.streamURL(), c.headers(), req)
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}

	out := make(chan core.StreamResult)
	go func() {
		defer resp.Body.Close()
		defer close(out)

		if resp.StatusCode != http.StatusOK {
			out <- core.StreamResult{Err: fmt.Errorf("gemini: %s", httpx.ErrorMessage(resp))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		for scanner.Scan() {
			line := bytes.TrimSpace(scanner.Bytes())
			if len(line) == 0 {
				continue
			}
			if bytes.HasPrefix(line, []byte("data:")) {
				line = bytes.TrimSpace(line[len("data:"):])
			}
			if len(line) == 0 || string(line) == "[DONE]" {
				continue
			}

			var chunk response
			if err := json.Unmarshal(line, &chunk); err != nil {
				out <- core.StreamResult{Err: fmt.Errorf("gemini: decode stream chunk: %w", err)}
				return
			}
			if chunk.Error != nil {
				out <- core.StreamResult{Err: fmt.Errorf("gemini: %s (%s)", chunk.Error.Message, chunk.Error.Status)}
				return
			}
			if len(chunk.Candidates) == 0 {
				continue
			}
			text := text_from_parts(chunk.Candidates[0].Content.Parts)
			if text != "" {
				out <- core.StreamResult{Response: text}
			}
		}
		if err := scanner.Err(); err != nil {
			out <- core.StreamResult{Err: err}
		}
	}()

	return out, nil
}
