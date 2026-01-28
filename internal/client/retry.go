package client

import (
	"bytes"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultMaxRetries = 3
	defaultMinRetryWait = 1 * time.Second
	defaultMaxRetryWait = 30 * time.Second
)

// RetryableTransport wraps an HTTP transport with retry logic
type RetryableTransport struct {
	Transport     http.RoundTripper
	MaxRetries    int
	MinRetryWait  time.Duration
	MaxRetryWait  time.Duration
}

// NewRetryableTransport creates a new retryable transport
func NewRetryableTransport(base http.RoundTripper) *RetryableTransport {
	if base == nil {
		base = http.DefaultTransport
	}

	return &RetryableTransport{
		Transport:    base,
		MaxRetries:   defaultMaxRetries,
		MinRetryWait: defaultMinRetryWait,
		MaxRetryWait: defaultMaxRetryWait,
	}
}

// RoundTrip executes a single HTTP transaction with retry logic
func (t *RetryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	// Read and store the request body if present (for retries)
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
	}

	for attempt := 0; attempt <= t.MaxRetries; attempt++ {
		// Reset the request body for each attempt
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Attempt the request
		resp, err = t.Transport.RoundTrip(req)

		// Check if we should retry
		if !t.shouldRetry(resp, err, attempt) {
			break
		}

		// Calculate backoff duration
		waitDuration := t.calculateBackoff(attempt)

		tflog.Debug(req.Context(), "Retrying request after error", map[string]interface{}{
			"attempt":      attempt + 1,
			"max_retries":  t.MaxRetries,
			"wait_seconds": waitDuration.Seconds(),
			"method":       req.Method,
			"url":          req.URL.String(),
		})

		// Wait before retrying
		time.Sleep(waitDuration)

		// Close the previous response body if it exists
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}

	return resp, err
}

// shouldRetry determines if a request should be retried
func (t *RetryableTransport) shouldRetry(resp *http.Response, err error, attempt int) bool {
	// Don't retry if we've exhausted our attempts
	if attempt >= t.MaxRetries {
		return false
	}

	// Retry on network errors
	if err != nil {
		return true
	}

	// Retry on specific HTTP status codes
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusTooManyRequests: // 429
			return true
		case http.StatusServiceUnavailable: // 503
			return true
		case http.StatusGatewayTimeout: // 504
			return true
		case http.StatusBadGateway: // 502
			return true
		}
	}

	return false
}

// calculateBackoff calculates the backoff duration using exponential backoff
func (t *RetryableTransport) calculateBackoff(attempt int) time.Duration {
	// Calculate exponential backoff: min * (2 ^ attempt)
	backoff := float64(t.MinRetryWait) * math.Pow(2, float64(attempt))

	// Cap at maximum wait time
	if backoff > float64(t.MaxRetryWait) {
		backoff = float64(t.MaxRetryWait)
	}

	return time.Duration(backoff)
}
