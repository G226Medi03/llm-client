package opencode

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/G226Medi03/llm-client/core"
)

func TestGenerateOpenAI(t *testing.T) {
	var captured openaiRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openai/v1/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("x-opencode-session") == "" {
			t.Errorf("missing x-opencode-session")
		}
		if r.Header.Get("User-Agent") == "" {
			t.Errorf("missing User-Agent")
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Errorf("decode: %v", err)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello"}}]}`))
	}))
	defer server.Close()

	client, err := New(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "deepseek-v4.1-flash",
		API:     APIOpenAI,
		Options: core.Options{JSONMode: true},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	answer, err := client.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if answer != "hello" {
		t.Fatalf("answer = %q", answer)
	}
	if captured.Model != "deepseek-v4.1-flash" {
		t.Fatalf("model = %q", captured.Model)
	}
	if captured.ResponseFormat == nil || captured.ResponseFormat.Type != "json_object" {
		t.Fatalf("response format = %#v", captured.ResponseFormat)
	}
}

func TestGenerateResponses(t *testing.T) {
	var captured responsesRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/openai/v1/responses" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Errorf("decode: %v", err)
		}
		_, _ = w.Write([]byte(`{"output":[{"type":"message","content":[{"type":"output_text","text":"sol answer"}]}]}`))
	}))
	defer server.Close()

	client, err := New(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-6.1-sol",
		API:     APIResponses,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	answer, err := client.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if answer != "sol answer" {
		t.Fatalf("answer = %q", answer)
	}
	if captured.Model != "gpt-6.1-sol" {
		t.Fatalf("model = %q", captured.Model)
	}
}
