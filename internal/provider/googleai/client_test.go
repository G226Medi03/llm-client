package googleai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateFiltersThoughtParts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Fatalf("api key header = %q", r.Header.Get("x-goog-api-key"))
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		generation, _ := request["generationConfig"].(map[string]any)
		thinking, _ := generation["thinkingConfig"].(map[string]any)
		if thinking["thinkingLevel"] != "minimal" {
			t.Fatalf("thinkingConfig = %#v", thinking)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"","thought":true},{"text":"③"}]}}]}`))
	}))
	defer server.Close()

	client, err := New(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gemma-4-26b-a4b-it",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	output, err := client.Generate(context.Background(), "question")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if strings.TrimSpace(output) != "③" {
		t.Fatalf("output = %q", output)
	}
}

func TestRequiresAPIKey(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected missing API key error")
	}
}
