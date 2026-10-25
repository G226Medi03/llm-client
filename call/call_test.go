package call

import (
	"context"
	"testing"
	"time"

	llmclient "github.com/G226Medi03/llm-client"
)

type fakeClient struct {
	output string
}

func (f *fakeClient) Model() string { return "fake" }
func (f *fakeClient) Generate(context.Context, string) (string, error) {
	return f.output, nil
}
func (f *fakeClient) GenerateStream(context.Context, string) (<-chan llmclient.StreamResult, error) {
	ch := make(chan llmclient.StreamResult)
	close(ch)
	return ch, nil
}
func (f *fakeClient) ResetContext() {}
func (f *fakeClient) Clone() llmclient.Client {
	copy := *f
	return &copy
}

func TestGenerate(t *testing.T) {
	value, err := Generate(context.Background(), &fakeClient{output: "ok"}, "prompt", time.Second)
	if err != nil {
		t.Fatalf("Generate: %s", err)
	}
	if value != "ok" {
		t.Fatalf("value = %q", value)
	}
}
