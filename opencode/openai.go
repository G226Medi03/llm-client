package opencode

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/G226Medi03/llm-client/core"
	"github.com/G226Medi03/llm-client/internal/httpx"
)

type openaiRequest struct {
	Model          string                `json:"model"`
	Messages       []openaiMessage       `json:"messages"`
	Stream         bool                  `json:"stream,omitempty"`
	ResponseFormat *openaiResponseFormat `json:"response_format,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type openaiContentPart struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *openaiImageURL `json:"image_url,omitempty"`
}

type openaiImageURL struct {
	URL string `json:"url"`
}

type openaiResponseFormat struct {
	Type string `json:"type"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *openaiError `json:"error,omitempty"`
}

type openaiStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Error *openaiError `json:"error,omitempty"`
}

type openaiError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (c *Client) generate_openai(ctx context.Context, prompt string) (string, error) {
	return c.generate_openai_content(ctx, prompt, "")
}

func (c *Client) generate_openai_content(ctx context.Context, prompt, imagePath string) (string, error) {
	req, err := c.build_openai_request(prompt, imagePath, false)
	if err != nil {
		return "", err
	}
	resp, err := httpx.PostJSON(ctx, c.http, c.url("/openai/v1/chat/completions"), c.headers(), req)
	if err != nil {
		return "", c.errorf("%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.errorf("%s", httpx.ErrorMessage(resp))
	}

	var parsed openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", c.errorf("decode response: %w", err)
	}
	if parsed.Error != nil {
		return "", c.errorf("%s (%s)", parsed.Error.Message, parsed.Error.Type)
	}
	if len(parsed.Choices) == 0 {
		return "", c.errorf("empty response")
	}
	content := parsed.Choices[0].Message.Content
	c.remember(prompt, content)
	return content, nil
}

func (c *Client) generate_openai_stream(ctx context.Context, prompt string) (<-chan core.StreamResult, error) {
	req, err := c.build_openai_request(prompt, "", true)
	if err != nil {
		return nil, err
	}
	resp, err := httpx.PostJSON(ctx, c.http, c.url("/openai/v1/chat/completions"), c.headers(), req)
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
			var chunk openaiStreamChunk
			if err := json.Unmarshal(line, &chunk); err != nil {
				out <- core.StreamResult{Err: c.errorf("decode stream chunk: %w", err)}
				return
			}
			if chunk.Error != nil {
				out <- core.StreamResult{Err: c.errorf("%s (%s)", chunk.Error.Message, chunk.Error.Type)}
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

func (c *Client) build_openai_request(prompt, imagePath string, stream bool) (*openaiRequest, error) {
	content := any(prompt)
	if imagePath != "" {
		part, err := image_content_part(imagePath)
		if err != nil {
			return nil, c.errorf("%w", err)
		}
		content = []openaiContentPart{
			{Type: "text", Text: prompt},
			part,
		}
	}

	messages := make([]openaiMessage, 0, len(c.savedMessages)+1)
	messages = append(messages, c.savedMessages...)
	messages = append(messages, openaiMessage{Role: "user", Content: content})

	req := &openaiRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   stream,
	}
	if c.jsonMode {
		req.ResponseFormat = &openaiResponseFormat{Type: "json_object"}
	}
	return req, nil
}

func (c *Client) remember(prompt, response string) {
	c.savedMessages = append(c.savedMessages,
		openaiMessage{Role: "user", Content: prompt},
		openaiMessage{Role: "assistant", Content: response},
	)
}

func (c *Client) generate_openai_image(ctx context.Context, prompt, imagePath string) (string, error) {
	return c.generate_openai_content(ctx, prompt, imagePath)
}

var _ core.OcrClient = (*Client)(nil)

func (c *Client) AskOCR(ctx context.Context, imagePath string) string {
	text, err := c.generate_openai_image(ctx, c.ocrPrompt, imagePath)
	if err != nil {
		return ""
	}
	return text
}

func (c *Client) GenerateWithImage(ctx context.Context, prompt string, imagePath string) (string, error) {
	if c.api != APIOpenAI {
		return "", fmt.Errorf("opencode: image input is only supported in openai mode")
	}
	return c.generate_openai_image(ctx, prompt, imagePath)
}
