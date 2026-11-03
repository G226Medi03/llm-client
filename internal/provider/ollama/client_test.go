package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/G226Medi03/llm-client/internal/core"
)

func newTestClient(t *testing.T, handler http.HandlerFunc, opts core.Options) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c, err := New(Config{BaseURL: srv.URL, Model: "test-model", Options: opts})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestNewRequiresModel(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("expected error for missing model")
	}
}

func TestGenerate(t *testing.T) {
	var gotPath string
	var gotBody chatRequest
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(chatResponse{Message: chatMessage{Role: "assistant", Content: "hi there"}, Done: true})
	}, core.Options{})

	out, err := c.Generate(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if out != "hi there" {
		t.Fatalf("got %q", out)
	}
	if gotPath != "/api/chat" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotBody.Format != "" {
		t.Fatalf("format = %q, want empty", gotBody.Format)
	}
}

func TestJSONMode(t *testing.T) {
	var gotBody chatRequest
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(chatResponse{Message: chatMessage{Content: "{}"}, Done: true})
	}, core.Options{JSONMode: true})

	if _, err := c.Generate(context.Background(), "hello"); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if gotBody.Format != "json" {
		t.Fatalf("format = %q, want json", gotBody.Format)
	}
}

func TestGeneratePassesThinkAndMaxTokens(t *testing.T) {
	var gotBody chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(chatResponse{Message: chatMessage{Content: "ok"}, Done: true})
	}))
	defer srv.Close()

	think := false
	client, err := New(Config{
		BaseURL:         srv.URL,
		Model:           "gemma4:31b",
		Think:           &think,
		MaxOutputTokens: 128,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := client.Generate(context.Background(), "hello"); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if gotBody.Think == nil || *gotBody.Think {
		t.Fatalf("think = %#v, want false", gotBody.Think)
	}
	if gotBody.Options == nil || gotBody.Options.NumPredict != 128 {
		t.Fatalf("options = %#v, want num_predict=128", gotBody.Options)
	}
}

func TestGenerateStream(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		flusher, _ := w.(http.Flusher)
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":"he"},"done":false}`)
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":"llo"},"done":false}`)
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":""},"done":true}`)
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
