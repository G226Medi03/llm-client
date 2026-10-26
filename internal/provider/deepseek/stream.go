package deepseek

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/G226Medi03/llm-client/internal/core"
	"github.com/G226Medi03/llm-client/internal/httpx"
)

// GenerateStream sends a text prompt and streams the response chunks.
func (c *Client) GenerateStream(ctx context.Context, prompt string) (<-chan core.StreamResult, error) {
	req, err := c.buildRequest(prompt, "", true)
	if err != nil {
		return nil, err
	}

	resp, err := httpx.PostJSON(ctx, c.http, c.chatURL(), map[string]string{
		"Authorization": "Bearer " + c.apiKey,
	}, req)
	if err != nil {
		return nil, fmt.Errorf("deepseek: %w", err)
	}

	out := make(chan core.StreamResult)

	go func() {
		defer resp.Body.Close()
		defer close(out)

		if resp.StatusCode != http.StatusOK {
			out <- core.StreamResult{Err: fmt.Errorf("deepseek: %s", httpx.ErrorMessage(resp))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

		var full strings.Builder
		for scanner.Scan() {
			line := bytes.TrimSpace(scanner.Bytes())
			if len(line) == 0 {
				continue
			}
			if bytes.HasPrefix(line, []byte("data:")) {
				line = bytes.TrimSpace(line[len("data:"):])
			}
			if string(line) == "[DONE]" {
				break
			}

			var chunk streamResponse
			if err := json.Unmarshal(line, &chunk); err != nil {
				out <- core.StreamResult{Err: fmt.Errorf("deepseek: decode stream chunk: %w", err)}
				return
			}
			if chunk.Error != nil {
				out <- core.StreamResult{Err: fmt.Errorf("deepseek: %s (%s)", chunk.Error.Message, chunk.Error.Type)}
				return
			}
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				content := chunk.Choices[0].Delta.Content
				full.WriteString(content)
				out <- core.StreamResult{Response: content}
			}
		}
		if err := scanner.Err(); err != nil {
			out <- core.StreamResult{Err: err}
			return
		}

		c.remember(prompt, full.String())
	}()

	return out, nil
}
