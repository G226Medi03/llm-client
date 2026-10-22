package llmclient

import "testing"

func TestParseProvider(t *testing.T) {
	cases := map[string]Provider{
		"deepseek": ProviderDeepSeek,
		"Ollama":   ProviderOllama,
		"gemini":   ProviderGemini,
		" Gemini ": ProviderGemini,
		" ollama ": ProviderOllama,
		"weird":    ProviderDeepSeek,
		"":         ProviderDeepSeek,
	}
	for in, want := range cases {
		if got := ParseProvider(in); got != want {
			t.Errorf("ParseProvider(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNewUnknownProvider(t *testing.T) {
	if _, err := New(Config{Provider: "nope"}); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestNewDeepSeekRequiresKey(t *testing.T) {
	if _, err := New(Config{Provider: ProviderDeepSeek}); err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestNewGeminiRequiresKey(t *testing.T) {
	if _, err := New(Config{Provider: ProviderGemini}); err == nil {
		t.Fatal("expected error for missing API key")
	}
}

func TestNewGemini(t *testing.T) {
	client, err := New(Config{
		Provider:      ProviderGemini,
		APIKey:        "test-key",
		Model:         "gemini-4-argon",
		ThinkingLevel: "high",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if client.Model() != "gemini-4-argon" {
		t.Fatalf("model = %q", client.Model())
	}
}

func TestNewOllama(t *testing.T) {
	client, err := New(Config{Provider: ProviderOllama, Model: "gemma4:e4b"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if client.Model() != "gemma4:e4b" {
		t.Fatalf("model = %q", client.Model())
	}
}
