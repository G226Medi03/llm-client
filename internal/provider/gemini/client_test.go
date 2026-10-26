package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/G226Medi03/llm-client/internal/core"
)

func TestGenerateFiltersThoughts(t *testing.T) {
	var captured request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Errorf("missing api key header")
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"candidates": [{
				"content": {"parts": [
					{"text": "internal thought", "thought": true},
					{"text": "final answer"}
				]}
			}]
		}`))
	}))
	defer server.Close()

	client, err := New(Config{
		APIKey:        "test-key",
		BaseURL:       server.URL,
		ThinkingLevel: ThinkingHigh,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	got, err := client.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "final answer" {
		t.Fatalf("Generate = %q", got)
	}
	if captured.GenerationConfig.ThinkingConfig.ThinkingLevel != ThinkingHigh {
		t.Fatalf("thinkingLevel = %q", captured.GenerationConfig.ThinkingConfig.ThinkingLevel)
	}
	if captured.GenerationConfig.MaxOutputTokens != 32768 {
		t.Fatalf("maxOutputTokens = %d", captured.GenerationConfig.MaxOutputTokens)
	}
}

func TestGenerateJSONMode(t *testing.T) {
	var captured request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{}"}]}}]}`))
	}))
	defer server.Close()

	client, err := New(Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Options: core.Options{JSONMode: true},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := client.Generate(context.Background(), "prompt"); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if captured.GenerationConfig.ResponseMimeType != "application/json" {
		t.Fatalf("responseMimeType = %q", captured.GenerationConfig.ResponseMimeType)
	}
}

func TestGenerateStreamFiltersThoughts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"thought\",\"thought\":true}]}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"answer\"}]}}]}\n\n"))
	}))
	defer server.Close()

	client, err := New(Config{APIKey: "test-key", BaseURL: server.URL})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	stream, err := client.GenerateStream(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("GenerateStream: %v", err)
	}
	var got string
	for item := range stream {
		if item.Err != nil {
			t.Fatalf("stream error: %v", item.Err)
		}
		got += item.Response
	}
	if got != "answer" {
		t.Fatalf("stream = %q", got)
	}
}
