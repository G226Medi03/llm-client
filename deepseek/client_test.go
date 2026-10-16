package deepseek

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/G226Medi03/llmclient/core"
)

func newTestClient(t *testing.T, handler http.HandlerFunc, opts core.Options) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := New(Config{APIKey: "test-key", BaseURL: srv.URL, Options: opts})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func TestNewRequiresAPIKey(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestGenerate(t *testing.T) {
	var gotPath, gotAuth string
	var gotBody request
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		writeJSON(w, response{Choices: []choice{{Message: respMessage{Content: "hello"}}}})
	}, core.Options{})

	out, err := c.Generate(context.Background(), "hi")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out != "hello" {
		t.Fatalf("got %q, want %q", out, "hello")
	}
	if gotPath != "/chat/completions" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer test-key" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if gotBody.ResponseFormat != nil {
		t.Fatalf("unexpected response_format: %+v", gotBody.ResponseFormat)
	}
}

func TestJSONMode(t *testing.T) {
	var gotBody request
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		writeJSON(w, response{Choices: []choice{{Message: respMessage{Content: "{}"}}}})
	}, core.Options{JSONMode: true})

	if _, err := c.Generate(context.Background(), "hi"); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if gotBody.ResponseFormat == nil || gotBody.ResponseFormat.Type != "json_object" {
		t.Fatalf("response_format = %+v", gotBody.ResponseFormat)
	}
}

func TestCloneClearsHistory(t *testing.T) {
	var sizes []int
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body request
		_ = json.NewDecoder(r.Body).Decode(&body)
		sizes = append(sizes, len(body.Messages))
		writeJSON(w, response{Choices: []choice{{Message: respMessage{Content: "ok"}}}})
	}, core.Options{})

	ctx := context.Background()
	if _, err := c.Generate(ctx, "first"); err != nil {
		t.Fatal(err)
	}
	clone := c.Clone()
	if _, err := clone.Generate(ctx, "second"); err != nil {
		t.Fatal(err)
	}
	if len(sizes) != 2 {
		t.Fatalf("requests = %d", len(sizes))
	}
	if sizes[0] != 1 {
		t.Fatalf("first request messages = %d, want 1", sizes[0])
	}
	if sizes[1] != 1 {
		t.Fatalf("clone request messages = %d, want 1", sizes[1])
	}
}

func TestGenerateStream(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		flusher, _ := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"he\"}}]}\n\n")
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"llo\"}}]}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
		if flusher != nil {
			flusher.Flush()
		}
	}, core.Options{})

	ch, err := c.GenerateStream(context.Background(), "hi")
	if err != nil {
		t.Fatalf("GenerateStream: %v", err)
	}
	var got string
	for res := range ch {
		if res.Err != nil {
			t.Fatalf("stream error: %v", res.Err)
		}
		got += res.Response
	}
	if got != "hello" {
		t.Fatalf("got %q, want %q", got, "hello")
	}
}

func TestAskOCR(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "page.png")
	if err := os.WriteFile(imgPath, []byte("fake-image"), 0o600); err != nil {
		t.Fatal(err)
	}

	var gotBody request
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		writeJSON(w, response{Choices: []choice{{Message: respMessage{Content: "  extracted  "}}}})
	}, core.Options{})

	out := c.AskOCR(context.Background(), imgPath)
	if out != "extracted" {
		t.Fatalf("got %q, want %q", out, "extracted")
	}

	parts, ok := gotBody.Messages[0].Content.([]any)
	if !ok || len(parts) != 2 {
		t.Fatalf("content parts = %#v", gotBody.Messages[0].Content)
	}
	second := parts[1].(map[string]any)
	if second["type"] != "image_url" {
		t.Fatalf("second part type = %v", second["type"])
	}
}
