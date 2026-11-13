package call

import (
	"context"
	"errors"
	"testing"
)

func TestRetryableError(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{context.DeadlineExceeded, true},
		{errors.New("deepseek: decode response: context deadline exceeded"), true},
		{errors.New("temporary timeout"), true},
		{context.Canceled, false},
		{errors.New("HTTP 429 too many requests"), true},
		{errors.New("rate limit exceeded"), true},
		{errors.New("invalid API key"), false},
	}
	for _, item := range cases {
		if got := retryable_error(item.err); got != item.want {
			t.Fatalf("retryable_error(%q) = %v, want %v", item.err, got, item.want)
		}
	}
}
