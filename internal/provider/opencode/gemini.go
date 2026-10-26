package opencode

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/G226Medi03/llm-client/internal/core"
	"github.com/G226Medi03/llm-client/internal/httpx"
)

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text    string `json:"text,omitempty"`
	Thought bool   `json:"thought,omitempty"`
}

type geminiGenerationConfig struct {
	ThinkingConfig geminiThinkingConfig `json:"thinkingConfig"`
}

type geminiThinkingConfig struct {
	ThinkingLevel string `json:"thinkingLevel"`
}

type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *openaiError      `json:"error,omitempty"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason,omitempty"`
}

type geminiStreamEvent = geminiResponse

func (c *Client) generate_gemini(ctx context.Context, prompt string) (string, error) {
	req := c.build_gemini_request(prompt)
	resp, err := httpx.PostJSON(ctx, c.http, c.url("/google/v1beta/models/"+c.model+":generateContent"), c.headers(), req)
	if err != nil {
		return "", c.errorf("%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.errorf("%s", httpx.ErrorMessage(resp))
	}

	var parsed geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", c.errorf("decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", c.errorf("%s (%s)", parsed.Error.Message, parsed.Error.Type)
	}
	if len(parsed.Candidates) == 0 {
		return "", c.errorf("empty response")
	}
	return gemini_text(parsed.Candidates[0].Content.Parts), nil
}

func (c *Client) generate_gemini_stream(ctx context.Context, prompt string) (<-chan core.StreamResult, error) {
	req := c.build_gemini_request(prompt)
	resp, err := httpx.PostJSON(ctx, c.http, c.url("/google/v1beta/models/"+c.model+":streamGenerateContent?alt=sse"), c.headers(), req)
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
		for scanner.Scan() {
			line := bytes.TrimSpace(scanner.Bytes())
			if len(line) == 0 || !bytes.HasPrefix(line, []byte("data:")) {
				continue
			}
			line = bytes.TrimSpace(line[len("data:"):])
			if len(line) == 0 || string(line) == "[DONE]" {
				continue
			}

			var event geminiStreamEvent
			if err := json.Unmarshal(line, &event); err != nil {
				out <- core.StreamResult{Err: c.errorf("decode stream chunk: %w", err)}
				return
			}
			if event.Error != nil {
				out <- core.StreamResult{Err: c.errorf("%s (%s)", event.Error.Message, event.Error.Type)}
				return
			}
			if len(event.Candidates) == 0 {
				continue
			}
			text := gemini_text(event.Candidates[0].Content.Parts)
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

func (c *Client) build_gemini_request(prompt string) geminiRequest {
	return geminiRequest{
		Contents: []geminiContent{{
			Role:  "user",
			Parts: []geminiPart{{Text: prompt}},
		}},
		GenerationConfig: geminiGenerationConfig{
			ThinkingConfig: geminiThinkingConfig{ThinkingLevel: c.thinkingLevel},
		},
	}
}

func gemini_text(parts []geminiPart) string {
	var builder strings.Builder
	for _, part := range parts {
		if part.Thought || part.Text == "" {
			continue
		}
		builder.WriteString(part.Text)
	}
	return builder.String()
}
