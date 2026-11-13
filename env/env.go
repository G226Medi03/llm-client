// Package env provides environment helpers shared by LLM components.
package env

import (
	"os"
	"strings"
	"time"
)

// String returns a trimmed environment value or fallback.
func String(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

// Duration parses a positive duration environment value or returns fallback.
func Duration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
