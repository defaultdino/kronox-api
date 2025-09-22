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

func (c *client) SendRequestWithFormData(ctx context.Context, method, endpoint string, params map[string]string, formData string) (*http.Response, error) {
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

	req, err := http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(formData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Set form-specific headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(formData)))

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	if method == http.MethodPost {
		req.Header.Set("Referer", fullURL)
	}

	if sessionID := getSessionFromContext(ctx); sessionID != "" {
		req.AddCookie(&http.Cookie{
			Name:  "JSESSIONID",
			Value: sessionID,
		})
	}

	if method == http.MethodPost && strings.Contains(fullURL, "login_do.jsp") {
		originalCheckRedirect := c.httpClient.CheckRedirect
		c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
		defer func() {
			c.httpClient.CheckRedirect = originalCheckRedirect
		}()
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
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

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Always set JSON content type for this method
	req.Header.Set("Content-Type", "application/json")
	if body != "" {
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(reqBody)))
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	if method == http.MethodPost {
		req.Header.Set("Referer", fullURL)
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

func (c *client) GetCookieJar() http.CookieJar {
	return c.cookieJar
}

func (c *client) CopyCookiesFrom(other Client, schoolUrl string) error {
	otherJar := other.GetCookieJar()
	if otherJar == nil {
		return nil
	}

	parsedURL, err := url.Parse(schoolUrl)
	if err != nil {
		return fmt.Errorf("failed to parse school URL: %w", err)
	}

	cookies := otherJar.Cookies(parsedURL)

	c.cookieJar.SetCookies(parsedURL, cookies)

	return nil
}

func getSessionFromContext(ctx context.Context) string {
	if sessionID := ctx.Value("session_id"); sessionID != nil {
		if sid, ok := sessionID.(string); ok {
			return sid
		}
	}
	return ""
}
