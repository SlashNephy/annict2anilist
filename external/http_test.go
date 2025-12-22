package external

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryTransport_Handles429WithRetryAfter(t *testing.T) {
	// Track the number of requests made
	requestCount := 0
	retryAfterSeconds := 1

	// Create a test server that returns 429 on first request, then 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount == 1 {
			// First request: return 429 with Retry-After header
			w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limited"))
			return
		}
		// Second request: return 200
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer server.Close()

	// Create a retry transport
	transport := &retryTransport{
		base: http.DefaultTransport,
	}

	client := &http.Client{
		Transport: transport,
	}

	// Make a request
	startTime := time.Now()
	resp, err := client.Get(server.URL)
	endTime := time.Now()

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, requestCount, "Expected 2 requests (1 failed + 1 retry)")

	// Check that we waited at least the retry-after duration
	elapsed := endTime.Sub(startTime)
	assert.GreaterOrEqual(t, elapsed.Seconds(), float64(retryAfterSeconds),
		"Should have waited at least %d seconds", retryAfterSeconds)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "Success", string(body))
	resp.Body.Close()
}

func TestRetryTransport_Handles429WithoutRetryAfter(t *testing.T) {
	// Create a test server that returns 429 without Retry-After header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Rate limited"))
	}))
	defer server.Close()

	// Create a retry transport
	transport := &retryTransport{
		base: http.DefaultTransport,
	}

	client := &http.Client{
		Transport: transport,
	}

	// Make a request
	resp, err := client.Get(server.URL)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Should return 429 without retry")
	resp.Body.Close()
}

func TestRetryTransport_Handles429WithInvalidRetryAfter(t *testing.T) {
	// Create a test server that returns 429 with invalid Retry-After header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "invalid")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("Rate limited"))
	}))
	defer server.Close()

	// Create a retry transport
	transport := &retryTransport{
		base: http.DefaultTransport,
	}

	client := &http.Client{
		Transport: transport,
	}

	// Make a request
	resp, err := client.Get(server.URL)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "Should return 429 without retry")
	resp.Body.Close()
}

func TestRetryTransport_HandlesNon429Responses(t *testing.T) {
	// Create a test server that returns 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer server.Close()

	// Create a retry transport
	transport := &retryTransport{
		base: http.DefaultTransport,
	}

	client := &http.Client{
		Transport: transport,
	}

	// Make a request
	resp, err := client.Get(server.URL)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "Success", string(body))
	resp.Body.Close()
}

func TestRetryTransport_HandlesRequestWithBody(t *testing.T) {
	requestCount := 0
	var receivedBodies []string

	// Create a test server that returns 429 on first request, then 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		body, _ := io.ReadAll(r.Body)
		receivedBodies = append(receivedBodies, string(body))

		if requestCount == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a retry transport
	transport := &retryTransport{
		base: http.DefaultTransport,
	}

	client := &http.Client{
		Transport: transport,
	}

	// Make a POST request with body
	requestBody := "test request body"
	resp, err := client.Post(server.URL, "text/plain", bytes.NewBufferString(requestBody))

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, requestCount, "Expected 2 requests")
	assert.Equal(t, requestBody, receivedBodies[0], "First request body should match")
	assert.Equal(t, requestBody, receivedBodies[1], "Second request body should match")
	resp.Body.Close()
}

func TestRetryTransport_MultipleRetries(t *testing.T) {
	requestCount := 0
	maxRetries := 3

	// Create a test server that returns 429 multiple times
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount <= maxRetries {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}))
	defer server.Close()

	// Create a retry transport
	transport := &retryTransport{
		base: http.DefaultTransport,
	}

	client := &http.Client{
		Transport: transport,
	}

	// Make a request
	startTime := time.Now()
	resp, err := client.Get(server.URL)
	endTime := time.Now()

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, maxRetries+1, requestCount, "Expected %d requests", maxRetries+1)

	// Check that we waited at least the retry-after duration * number of retries
	elapsed := endTime.Sub(startTime)
	assert.GreaterOrEqual(t, elapsed.Seconds(), float64(maxRetries),
		"Should have waited at least %d seconds", maxRetries)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "Success", string(body))
	resp.Body.Close()
}
