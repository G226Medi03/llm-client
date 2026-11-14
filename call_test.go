package llmclient

import (
	"context"
	"testing"
	"time"
)

type fakeClient struct {
	output string
}

func (*fakeClient) is_LlmClient() {}

func (f *fakeClient) Model() string { return "fake" }
func (f *fakeClient) Generate(context.Context, string) (LlmResponse, error) {
	return LlmResponse(f.output), nil
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
	if value != LlmResponse("ok") {
		t.Fatalf("value = %q", value)
	}
}
