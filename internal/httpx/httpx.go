// Package httpx contains small helpers shared by llmclient providers.
package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PostJSON marshals payload and performs an HTTP POST with a JSON body.
// The caller owns the returned response body.
func PostJSON(ctx context.Context, hc *http.Client, url string, headers map[string]string, payload any) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return hc.Do(req)
}

// ErrorMessage reads a short body from a failed response for error reporting.
func ErrorMessage(resp *http.Response) string {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Sprintf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}
