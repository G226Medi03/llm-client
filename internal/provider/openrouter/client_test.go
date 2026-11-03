package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("HTTP-Referer") != "https://example.test" {
			t.Fatalf("referer = %q", r.Header.Get("HTTP-Referer"))
		}
		if r.Header.Get("X-OpenRouter-Title") != "MorbesGoDB" {
			t.Fatalf("title = %q", r.Header.Get("X-OpenRouter-Title"))
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request["model"] != "google/gemma-4-26b-a4b-it:free" {
			t.Fatalf("model = %v", request["model"])
		}
		if _, ok := request["provider"]; ok {
			t.Fatalf("provider should be omitted when unset: %v", request["provider"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"③"}}]}`))
	}))
	defer server.Close()

	client, err := New(Config{
		APIKey:      "test-key",
		BaseURL:     server.URL,
		Model:       "google/gemma-4-26b-a4b-it:free",
		HTTPReferer: "https://example.test",
		Title:       "MorbesGoDB",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	output, err := client.Generate(context.Background(), "question")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if output != "③" {
		t.Fatalf("output = %q", output)
	}
}

func TestGenerateSendsProviderPreferences(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Provider *struct {
				Only           []string `json:"only"`
				AllowFallbacks bool     `json:"allow_fallbacks"`
			} `json:"provider"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.Provider == nil {
			t.Fatal("provider is missing")
		}
		if len(request.Provider.Only) != 1 || request.Provider.Only[0] != "deepinfra" {
			t.Fatalf("provider.only = %#v", request.Provider.Only)
		}
		if request.Provider.AllowFallbacks {
			t.Fatal("allow_fallbacks = true, want false")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"③"}}]}`))
	}))
	defer server.Close()

	allowFallbacks := false
	client, err := New(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "google/gemma-4-31b-it",
		Provider: &ProviderPreferences{
			Only:           []string{"deepinfra"},
			AllowFallbacks: &allowFallbacks,
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	output, err := client.Generate(context.Background(), "question")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if output != "③" {
		t.Fatalf("output = %q", output)
	}
}

func TestRequiresAPIKey(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected missing API key error")
	}
}
