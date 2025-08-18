package kronox

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type client struct {
	httpClient *http.Client
	cookieJar  http.CookieJar
}

func NewClient(httpClient *http.Client) Client {
	jar, _ := cookiejar.New(nil)
	httpClient.Jar = jar

	return &client{
		httpClient: httpClient,
		cookieJar:  jar,
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

	var reqBody string
	if body != "" {
		reqBody = body
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "KronoxAPI/1.0")
	if method == http.MethodPost && body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req.Header.Set("Accept", "application/xml, text/html")
	}

	if sessionID := getSessionFromContext(ctx); sessionID != "" {
		req.AddCookie(&http.Cookie{
			Name:  "JSESSIONID",
			Value: sessionID,
		})
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

func getSessionFromContext(ctx context.Context) string {
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		if sid, ok := sessionID.(string); ok {
			return sid
		}
	}
	return ""
}
