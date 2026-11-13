// Package call provides the application-level timeout policy for LLM calls.
package call

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/G226Medi03/llm-client/env"
)

const (
	default_request_timeout = 5 * time.Minute
	default_request_retries = 1
	default_retry_backoff   = 2 * time.Second
	default_max_concurrent  = 8
	max_allowed_concurrency = 32
)

var (
	llm_slots_once sync.Once
	llm_slots      chan struct{}
)

// LLM_MAX_CONCURRENT caps simultaneous LLM HTTP requests. Multiple solve
// workers can each have one or two extraction calls in flight, so this guard
// prevents a larger worker pool from overwhelming the provider.
func LLM_MAX_CONCURRENT() int {
	value := int_env("LLM_MAX_CONCURRENT", default_max_concurrent)
	if value < 1 {
		return 1
	}
	if value > max_allowed_concurrency {
		return max_allowed_concurrency
	}
	return value
}

func acquire_llm_slot(ctx context.Context) error {
	llm_slots_once.Do(func() {
		llm_slots = make(chan struct{}, LLM_MAX_CONCURRENT())
	})
	select {
	case llm_slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func release_llm_slot() {
	<-llm_slots
}

// LLM_REQUEST_TIMEOUT returns the per-call timeout configured by LLM_REQUEST_TIMEOUT.
func LLM_REQUEST_TIMEOUT() time.Duration {
	return env.Duration("LLM_REQUEST_TIMEOUT", default_request_timeout)
}

// LLM_REQUEST_RETRIES returns the number of retries after the initial attempt.
func LLM_REQUEST_RETRIES() int {
	return int_env("LLM_REQUEST_RETRIES", default_request_retries)
}

func generateWithPolicy(ctx context.Context, generate GenerateFunc, prompt string) (string, error) {
	if err := acquire_llm_slot(ctx); err != nil {
		return "", err
	}
	defer release_llm_slot()
	attempts := LLM_REQUEST_RETRIES() + 1
	var last_err error
	for attempt := 0; attempt < attempts; attempt++ {
		if ctx_err := ctx.Err(); ctx_err != nil {
			return "", ctx_err
		}
		output, err := generate(ctx, prompt)
		if err == nil {
			return output, nil
		}
		last_err = err
		if attempt+1 >= attempts || !retryable_error(err) {
			return "", err
		}
		select {
		case <-time.After(default_retry_backoff):
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	return "", last_err
}

func retryable_error(err error) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	message := strings.ToLower(err.Error())
	for _, token := range []string{
		"context deadline exceeded",
		"timeout",
		"timed out",
		"connection reset",
		"connection refused",
		"temporarily unavailable",
		"unexpected eof",
		"bad gateway",
		"service unavailable",
		"gateway timeout",
		"429",
		"rate limit",
		"rate_limit",
		"too many requests",
	} {
		if strings.Contains(message, token) {
			return true
		}
	}
	return false
}

func int_env(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return fallback
	}
	return parsed
}
