package kronox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) Client {
	return &client{
		httpClient: httpClient,
	}
}

func (c *client) SendRequest(ctx context.Context, method, endpoint string, params map[string]string) (*http.Response, error) {
	return c.SendRequestWithBody(ctx, method, endpoint, params, "")
}

func (c *client) SendRequestWithBody(ctx context.Context, method, endpoint string, params map[string]string, body string) (*http.Response, error) {
	fullURL := endpoint

	if method == http.MethodGet && len(params) > 0 {
		values := url.Values{}
		for key, value := range params {
			values.Add(key, value)
		}

		separator := "?"
		if strings.Contains(fullURL, "?") {
			separator = "&"
		}
		fullURL += separator + values.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	req.Header.Set("Content-Type", "application/json")
	if body != "" {
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	if method == http.MethodPost {
		req.Header.Set("Referer", fullURL)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}
