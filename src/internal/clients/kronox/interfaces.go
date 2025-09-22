package kronox

import (
	"context"
	"net/http"
)

type Client interface {
	SendRequest(ctx context.Context, method, endpoint string, params map[string]string) (*http.Response, error)
	SendRequestWithBody(ctx context.Context, method, endpoint string, params map[string]string, body string) (*http.Response, error)
	SendRequestWithFormData(ctx context.Context, method, endpoint string, params map[string]string, formData string) (*http.Response, error)
	CopyCookiesFrom(other Client, schoolUrl string) error
	GetCookieJar() http.CookieJar
}
