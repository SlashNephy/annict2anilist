package external

import (
	"log/slog"
	"net/http"
	"time"
)

func NewHttpClient() *http.Client {
	return &http.Client{
		Transport: &loggingTransport{},
		Timeout:   30 * time.Second,
	}
}

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
