package opencode

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/G226Medi03/llm-client/core"
	"github.com/G226Medi03/llm-client/internal/httpx"
)

type responsesRequest struct {
	Model  string `json:"model"`
	Input  string `json:"input"`
	Stream bool   `json:"stream,omitempty"`
}

type responsesResponse struct {
	OutputText string            `json:"output_text"`
	Output     []responsesOutput `json:"output"`
	Error      *openaiError      `json:"error,omitempty"`
}

type responsesOutput struct {
	Type    string             `json:"type"`
	Content []responsesContent `json:"content"`
}

type responsesContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responsesStreamEvent struct {
	Type     string             `json:"type"`
	Delta    string             `json:"delta"`
	Response *responsesResponse `json:"response,omitempty"`
	Error    *openaiError       `json:"error,omitempty"`
}

func (c *Client) generate_responses(ctx context.Context, prompt string) (string, error) {
	req := responsesRequest{Model: c.model, Input: prompt}
	resp, err := httpx.PostJSON(ctx, c.http, c.url("/openai/v1/responses"), c.headers(), req)
	if err != nil {
		return "", c.errorf("%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.errorf("%s", httpx.ErrorMessage(resp))
	}

	var parsed responsesResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", c.errorf("decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", c.errorf("%s (%s)", parsed.Error.Message, parsed.Error.Type)
	}
	text := responses_text(parsed)
	if text == "" {
		return "", c.errorf("empty response")
	}
	return text, nil
}

func (c *Client) generate_responses_stream(ctx context.Context, prompt string) (<-chan core.StreamResult, error) {
	req := responsesRequest{Model: c.model, Input: prompt, Stream: true}
	resp, err := httpx.PostJSON(ctx, c.http, c.url("/openai/v1/responses"), c.headers(), req)
	if err != nil {
		return nil, c.errorf("%w", err)
	}

	out := make(chan core.StreamResult)
	go func() {
		defer resp.Body.Close()
		defer close(out)

		if resp.StatusCode != http.StatusOK {
			out <- core.StreamResult{Err: c.errorf("%s", httpx.ErrorMessage(resp))}
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
		var full strings.Builder
		sent := false
		for scanner.Scan() {
			line := bytes.TrimSpace(scanner.Bytes())
			if len(line) == 0 || !bytes.HasPrefix(line, []byte("data:")) {
				continue
			}
			line = bytes.TrimSpace(line[len("data:"):])
			if len(line) == 0 || string(line) == "[DONE]" {
				continue
			}

			var event responsesStreamEvent
			if err := json.Unmarshal(line, &event); err != nil {
				out <- core.StreamResult{Err: c.errorf("decode stream chunk: %w", err)}
				return
			}
			if event.Error != nil {
				out <- core.StreamResult{Err: c.errorf("%s (%s)", event.Error.Message, event.Error.Type)}
				return
			}
			if event.Type == "response.output_text.delta" && event.Delta != "" {
				sent = true
				full.WriteString(event.Delta)
				out <- core.StreamResult{Response: event.Delta}
				continue
			}
			if event.Type == "response.completed" && event.Response != nil && !sent {
				text := responses_text(*event.Response)
				if text != "" {
					full.WriteString(text)
					out <- core.StreamResult{Response: text}
					sent = true
				}
			}
		}
		if err := scanner.Err(); err != nil {
			out <- core.StreamResult{Err: err}
			return
		}
		_ = full.String()
	}()
	return out, nil
}

func responses_text(response responsesResponse) string {
	if strings.TrimSpace(response.OutputText) != "" {
		return response.OutputText
	}
	var builder strings.Builder
	for _, item := range response.Output {
		for _, content := range item.Content {
			if content.Text != "" {
				builder.WriteString(content.Text)
			}
		}
	}
	return builder.String()
}
