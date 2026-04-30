package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/defaultdino/kronox-api/internal/app"
)

type ProgrammeService struct {
	app *app.App
}

func NewProgrammeService(app *app.App) *ProgrammeService {
	return &ProgrammeService{app: app}
}

func (s *ProgrammeService) GetProgrammes(ctx context.Context, schoolUrl string, searchQuery string) (string, error) {
	endpoint := fmt.Sprintf("%s/ajax/ajax_sokResurser.jsp", strings.TrimSuffix(schoolUrl, "/"))

	params := map[string]string{
		"sokord":         searchQuery,
		"startDatum":     "idag",
		"slutDatum":      "",
		"intervallTyp":   "m",
		"intervallAntal": "6",
	}

	client := s.app.NewKronoxClient()

	response, err := client.SendRequest(ctx, http.MethodGet, endpoint, params)
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
