package external

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

func NewHttpClient() *http.Client {
	return &http.Client{
		Transport: &retryTransport{
			base: &loggingTransport{},
		},
		Timeout: 15 * time.Second,
	}
}

// retryTransport handles 429 rate limit errors with Retry-After header
type retryTransport struct {
	base http.RoundTripper
}

const (
	maxRetries       = 5
	maxRetryAfterSec = 300 // 5 minutes
)

func (t *retryTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	// Read and store the request body so we can retry if needed
	var bodyBytes []byte
	if request.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(request.Body)
		if err != nil {
			return nil, err
		}
		request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	retryCount := 0
	for {
		// Restore the request body for retry attempts
		if bodyBytes != nil {
			request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		response, err := t.base.RoundTrip(request)
		if err != nil {
			return nil, err
		}

		// Check if we got a 429 rate limit error
		if response.StatusCode == http.StatusTooManyRequests {
			// Check if we've exceeded max retries
			if retryCount >= maxRetries {
				slog.Warn("max retries exceeded for rate limited request",
					slog.String("url", request.URL.String()),
					slog.Int("retry_count", retryCount),
				)
				return response, nil
			}

			// Get the Retry-After header
			retryAfterStr := response.Header.Get("Retry-After")
			if retryAfterStr == "" {
				// No Retry-After header, return the error response
				return response, nil
			}

			// Parse Retry-After header (in seconds)
			// Note: This implementation only supports delta-seconds format, not HTTP-date format
			retryAfter, err := strconv.Atoi(retryAfterStr)
			if err != nil {
				// If parsing fails, return the error response
				slog.Warn("failed to parse Retry-After header",
					slog.String("value", retryAfterStr),
					slog.String("error", err.Error()),
				)
				return response, nil
			}

			// Validate retry-after value to prevent excessive waiting
			if retryAfter > maxRetryAfterSec {
				slog.Warn("retry-after exceeds maximum, returning error",
					slog.String("url", request.URL.String()),
					slog.Int("retry_after_seconds", retryAfter),
					slog.Int("max_seconds", maxRetryAfterSec),
				)
				return response, nil
			}

			// Close the response body before retrying
			response.Body.Close()

			// Wait for the specified time
			waitDuration := time.Duration(retryAfter) * time.Second
			slog.Info("rate limited, retrying after delay",
				slog.String("url", request.URL.String()),
				slog.Int("retry_after_seconds", retryAfter),
				slog.Int("retry_attempt", retryCount+1),
			)
			time.Sleep(waitDuration)

			retryCount++
			// Retry the request
			continue
		}

		// Not a 429 error, return the response
		return response, nil
	}
}

var _ http.RoundTripper = (*retryTransport)(nil)

type loggingTransport struct{}

const userAgent = "annict2anilist/1.0 (+https://github.com/SlashNephy/annict2anilist)"

func (*loggingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	slog.Debug("http request",
		slog.String("method", request.Method),
		slog.String("url", request.URL.String()),
	)

	request.Header.Set("User-Agent", userAgent)

	t1 := time.Now()
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	t2 := time.Now()

	slog.Debug("http response",
		slog.String("status", response.Status),
		slog.String("method", response.Request.Method),
		slog.String("url", response.Request.URL.String()),
		slog.String("duration", t2.Sub(t1).String()),
	)

	return response, nil
}

var _ http.RoundTripper = (*loggingTransport)(nil)
