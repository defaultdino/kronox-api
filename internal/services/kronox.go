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

type KronoxService struct {
	app *app.App
}

func NewKronoxService(app *app.App) *KronoxService {
	return &KronoxService{app: app}
}

func (s *KronoxService) GetSchedule(ctx context.Context, schoolURL string, scheduleIDs []string, language *string, startDate *time.Time) (string, error) {
	parsedDate := "idag"
	if startDate != nil {
		parsedDate = startDate.Format("2006-01-02")
	}

	parsedLang := "SV"
	if language != nil {
		parsedLang = *language
	}

	endpoint := fmt.Sprintf("%s/setup/jsp/SchemaXML.jsp", strings.TrimSuffix(schoolURL, "/"))

	params := map[string]string{
		"startDatum":     parsedDate,
		"intervallTyp":   "m",
		"intervallAntal": "6",
		"sprak":          parsedLang,
		"sokMedAND":      "false",
		"forklaringar":   "true",
		"resurser":       strings.Join(scheduleIDs, ","),
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
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

func (s *KronoxService) GetProgrammes(ctx context.Context, schoolURL string, searchQuery string) (string, error) {
	endpoint := fmt.Sprintf("%s/ajax/ajax_sokResurser.jsp", strings.TrimSuffix(schoolURL, "/"))

	params := map[string]string{
		"sokord":         searchQuery,
		"startDatum":     "idag",
		"slutDatum":      "",
		"intervallTyp":   "m",
		"intervallAntal": "6",
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return "", fmt.Errorf("failed to fetch programmes: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
