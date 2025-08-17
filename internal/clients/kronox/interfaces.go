package kronox

import (
	"context"
	"net/http"

	"github.com/tumble-for-kronox/kronox-api/pkg/models"
)

type Client interface {
	// endpoint should be a full URL (e.g., "https://schema.hkr.se/setup/jsp/SchemaXML.jsp")
	SendRequest(ctx context.Context, method, endpoint string, params map[string]string) (*http.Response, error)
}

type Logger interface {
	Info(msg string, fields ...any)
	Error(msg string, err error, fields ...any)
	Debug(msg string, fields ...any)
}

type ScheduleRepository interface {
	GetSchedule(ctx context.Context, scheduleID string) (*models.Schedule, error)
	SaveSchedule(ctx context.Context, schedule *models.Schedule) error
}
