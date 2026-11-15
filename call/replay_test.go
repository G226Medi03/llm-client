package call

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewReplayPromptMatch(t *testing.T) {
	replay := NewReplay(ReplayOptions{
		Calls: []ReplayCall{
			{Stage: "status_extract", Prompt: "prompt", Output: "stored"},
		},
	})
	got, err := replay.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "stored" {
		t.Fatalf("output = %q", got)
	}
}

func TestNewReplayStageFallback(t *testing.T) {
	replay := NewReplay(ReplayOptions{
		Calls: []ReplayCall{
			{Stage: "status_extract", Prompt: "original", Output: "stored"},
		},
		StageOf: func(string) string { return "status_extract" },
	})
	got, err := replay.Generate(context.Background(), "new prompt")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "stored" {
		t.Fatalf("output = %q", got)
	}
}

func TestNewReplayLiveStage(t *testing.T) {
	calls := 0
	live := NewFunc("live-model", func(context.Context, string) (string, error) {
		calls++
		return "live", nil
	})
	replay := NewReplay(ReplayOptions{
		Calls: []ReplayCall{
			{Stage: "status_extract", Prompt: "prompt", Output: "stored"},
		},
		Live:       live,
		LiveStages: map[string]bool{"status_extract": true},
		StageOf:    func(string) string { return "status_extract" },
	})
	got, err := replay.Generate(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "live" {
		t.Fatalf("output = %q", got)
	}
	if calls != 1 {
		t.Fatalf("live calls = %d", calls)
	}
	if replay.Model() != "live-model" {
		t.Fatalf("model = %q", replay.Model())
	}
}

func TestNewReplayLiveWithoutClient(t *testing.T) {
	replay := NewReplay(ReplayOptions{
		Calls: []ReplayCall{
			{Stage: "status_extract", Prompt: "prompt", Output: "stored"},
		},
		LiveStages: map[string]bool{"status_extract": true},
		StageOf:    func(string) string { return "status_extract" },
	})
	if _, err := replay.Generate(context.Background(), "prompt"); err == nil {
		t.Fatal("expected live client error")
	}
}

func TestNewReplayMissingOutput(t *testing.T) {
	replay := NewReplay(ReplayOptions{})
	if _, err := replay.Generate(context.Background(), "prompt"); err == nil {
		t.Fatal("expected missing output error")
	}
	if replay.Model() != "replay" {
		t.Fatalf("model = %q", replay.Model())
	}
}

func TestNewReplayDoesNotNestConcurrencySlots(t *testing.T) {
	live := NewFunc("live", func(context.Context, string) (string, error) {
		return "live", nil
	})
	replay := NewReplay(ReplayOptions{
		Calls: []ReplayCall{
			{Stage: "live", Prompt: "prompt", Output: "stored"},
		},
		Live:       live,
		LiveStages: map[string]bool{"live": true},
		StageOf:    func(string) string { return "live" },
	})

	const workers = 32
	var wg sync.WaitGroup
	done := make(chan struct{})
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := replay.Generate(context.Background(), "prompt"); err != nil {
				t.Errorf("Generate: %v", err)
			}
		}()
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("replay live calls deadlocked")
	}
}
