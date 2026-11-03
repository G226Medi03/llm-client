package novita

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
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request["model"] != "google/gemma-4-31b-it" {
			t.Fatalf("model = %v", request["model"])
		}
		if request["max_tokens"] != float64(128) {
			t.Fatalf("max_tokens = %v", request["max_tokens"])
		}
		if request["stream"] != false {
			t.Fatalf("stream = %v", request["stream"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"③"}}]}`))
	}))
	defer server.Close()

	client, err := New(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "google/gemma-4-31b-it",
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

func TestErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer server.Close()

	client, err := New(Config{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := client.Generate(context.Background(), "question"); err == nil {
		t.Fatal("expected status error")
	}
}

func TestEmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer server.Close()

	client, err := New(Config{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := client.Generate(context.Background(), "question"); err == nil {
		t.Fatal("expected empty response error")
	}
}
