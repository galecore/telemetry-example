package echohttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client struct {
	endpoint string

	client *http.Client // we expect the http client to be pre-configured with tracing and metrics
}

func NewWithClient(endpoint string, c *http.Client) *Client {
	return &Client{
		endpoint: endpoint,
		client:   c,
	}
}

func New(endpoint string, timeout time.Duration) *Client {
	transport := otelhttp.NewTransport(
		http.DefaultTransport,
		// otelhttp.WithSpanNameFormatter() could be passed, default span formatter is used otherwise
		// otelhttp.WithPropagators() could be passed, global propagators are used otherwise
	)
	return NewWithClient(endpoint, &http.Client{
		Transport: transport,
		Timeout:   timeout,
	})
}

func (c *Client) Echo(ctx context.Context, message string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/echo", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	queryArgs := request.URL.Query()
	queryArgs.Set("message", message)
	request.URL.RawQuery = queryArgs.Encode()

	response, err := c.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	result, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(result), nil
}
