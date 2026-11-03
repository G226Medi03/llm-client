package nvidia

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
		if request["model"] != "test-model" {
			t.Fatalf("model = %v", request["model"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"③"}}]}`))
	}))
	defer server.Close()

	client, err := New(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "test-model",
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
