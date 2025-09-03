package kronox

import (
	"context"
	"net/http"
)

type Client interface {
	SendRequest(ctx context.Context, method, endpoint string, params map[string]string) (*http.Response, error)
	SendRequestWithBody(ctx context.Context, method, endpoint string, params map[string]string, body string) (*http.Response, error)
	ResetCookieJar() error
}
