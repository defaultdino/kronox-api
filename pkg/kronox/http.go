package kronox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func doRequest(ctx context.Context, hc *http.Client, userAgent, endpoint string, params map[string]string) (string, error) {
	values := url.Values{}
	for k, v := range params {
		values.Add(k, v)
	}
	sep := "?"
	if strings.Contains(endpoint, "?") {
		sep = "&"
	}
	fullURL := endpoint + sep + values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("kronox request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return "", &APIError{StatusCode: resp.StatusCode, Status: resp.Status}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}
	return string(body), nil
}
