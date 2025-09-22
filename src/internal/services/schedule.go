package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tumble-for-kronox/kronox-api/internal/app"
)

type ScheduleService struct {
	app *app.App
}

func NewScheduleService(app *app.App) *ScheduleService {
	return &ScheduleService{app: app}
}

func (s *ScheduleService) GetScheduleEvents(ctx context.Context, schoolUrl string, scheduleIDs []string, startDate *time.Time) (string, error) {
	parsedDate := "idag"
	if startDate != nil {
		parsedDate = startDate.Format("2006-01-02")
	}

	endpoint := fmt.Sprintf("%s/setup/jsp/SchemaXML.jsp", strings.TrimSuffix(schoolUrl, "/"))

	params := map[string]string{
		"startDatum":     parsedDate,
		"intervallTyp":   "m",
		"intervallAntal": "6", // 6 months ahead (maximum)
		"sprak":          "EN",
		"sokMedAND":      "false",
		"forklaringar":   "true",
		"resurser":       strings.Join(scheduleIDs, ","),
	}

	client := s.app.NewKronoxClient()

	response, err := client.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return "", fmt.Errorf("failed to fetch schedule: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
