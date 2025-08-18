package kronox

import (
	"context"
	"net/http"

	"github.com/tumble-for-kronox/kronox-api/pkg/models"
)

type Client interface {
	SendRequest(ctx context.Context, method, endpoint string, params map[string]string) (*http.Response, error)
	SendRequestWithBody(ctx context.Context, method, endpoint string, params map[string]string, body string) (*http.Response, error)
}

type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, err error, fields ...any)
	Debug(msg string, fields ...any)
}

type ScheduleRepository interface {
	GetScheduleEvents(ctx context.Context, scheduleID string) ([]*models.Event, error)
	SaveScheduleEvents(ctx context.Context, schedule []*models.Event) error
}
