package llmclient

import (
	"context"
	"testing"
	"time"

	"github.com/G226Medi03/llm-client/internal/core"
)

type fakeClient struct {
	core.Sealed
	output string
}

func (f *fakeClient) Model() string { return "fake" }
func (f *fakeClient) Generate(context.Context, string) (string, error) {
	return f.output, nil
}
func (f *fakeClient) GenerateStream(context.Context, string) (<-chan StreamResult, error) {
	ch := make(chan StreamResult)
	close(ch)
	return ch, nil
}
func (f *fakeClient) ResetContext() {}
func (f *fakeClient) Clone() LlmClient {
	copy := *f
	return &copy
}

func TestGenerateIsolated(t *testing.T) {
	value, err := GenerateIsolated(context.Background(), &fakeClient{output: "ok"}, "prompt", time.Second)
	if err != nil {
		t.Fatalf("GenerateIsolated: %s", err)
	}
	if value != "ok" {
		t.Fatalf("value = %q", value)
	}
}
