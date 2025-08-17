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

	req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "KronoxAPI/1.0")
	req.Header.Set("Accept", "application/xml, text/html")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}
