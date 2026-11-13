package call

import (
	"context"
	"testing"
)

func TestNewFuncClient(t *testing.T) {
	client := NewFunc("test-model", func(ctx context.Context, prompt string) (string, error) {
		return "output:" + prompt, nil
	})
	if client == nil {
		t.Fatal("client is nil")
	}
	if client.Model() != "test-model" {
		t.Fatalf("model = %q", client.Model())
	}
	output, err := client.Generate(context.Background(), "hello")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if output != "output:hello" {
		t.Fatalf("output = %q", output)
	}
	clone := client.Clone()
	if clone == nil || clone.Model() != "test-model" {
		t.Fatalf("clone = %#v", clone)
	}
}
