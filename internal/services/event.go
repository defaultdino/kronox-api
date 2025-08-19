package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tumble-for-kronox/kronox-api/internal/app"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
)

type EventService struct {
	app            *app.App
	sessionService *SessionService
	parserService  parsers.ParserService
}

func NewEventService(app *app.App, sessionService *SessionService, parserService parsers.ParserService) *EventService {
	return &EventService{
		app:            app,
		sessionService: sessionService,
		parserService:  parserService,
	}
}

func (s *EventService) GetUserEvents(ctx context.Context, schoolUrl, sessionID string) (*parsers.EventsResponse, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	if err := s.sessionService.SetSessionLanguage(ctx, schoolUrl, sessionID); err != nil {
		return nil, fmt.Errorf("failed to set session language: %w", err)
	}

	endpoint := fmt.Sprintf("%s/aktivitetsanmalan.jsp", strings.TrimSuffix(schoolUrl, "/"))

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("failed to get user events: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	htmlContent := string(body)
	if len(htmlContent) == 0 {
		return nil, fmt.Errorf("received empty response from user events endpoint")
	}

	if strings.Contains(strings.ToLower(htmlContent), "användarnamn:") &&
		strings.Contains(strings.ToLower(htmlContent), "lösenord:") {
		return nil, fmt.Errorf("session expired - redirected to login page")
	}

	events, err := s.parserService.ParseUserEvents(htmlContent)

	if err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}

	return events, nil
}

func (s *EventService) RegisterUserEvent(ctx context.Context, schoolUrl, sessionID, userEventID string) error {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	if err := s.sessionService.SetSessionLanguage(ctx, schoolUrl, sessionID); err != nil {
		return fmt.Errorf("failed to set session language: %w", err)
	}

	endpoint := fmt.Sprintf("%s/ajax/ajax_aktivitetsanmalan.jsp", strings.TrimSuffix(schoolUrl, "/"))
	params := map[string]string{
		"op":                     "anmal",
		"aktivitetsTillfallesId": userEventID,
		"ort":                    "",
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to register for event: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register for event, status code: %d", response.StatusCode)
	}

	return nil
}

func (s *EventService) UnregisterUserEvent(ctx context.Context, schoolUrl, sessionID, userEventID string) error {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	if err := s.sessionService.SetSessionLanguage(ctx, schoolUrl, sessionID); err != nil {
		return fmt.Errorf("failed to set session language: %w", err)
	}

	endpoint := fmt.Sprintf("%s/ajax/ajax_aktivitetsanmalan.jsp", strings.TrimSuffix(schoolUrl, "/"))
	params := map[string]string{
		"op":                   "avanmal",
		"deltagarMojlighetsId": userEventID,
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to unregister from event: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to unregister from event, status code: %d", response.StatusCode)
	}

	return nil
}

func (s *EventService) AddEventSupport(ctx context.Context, schoolUrl, sessionID, participatorID, supportID string) error {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	if err := s.sessionService.SetSessionLanguage(ctx, schoolUrl, sessionID); err != nil {
		return fmt.Errorf("failed to set session language: %w", err)
	}

	endpoint := fmt.Sprintf("%s/ajax/ajax_aktivitetsanmalan.jsp", strings.TrimSuffix(schoolUrl, "/"))
	params := map[string]string{
		"op":         "laggTillStod",
		"stodId":     supportID,
		"deltagarId": participatorID,
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to add event support: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add event support, status code: %d", response.StatusCode)
	}

	return nil
}

func (s *EventService) RemoveEventSupport(ctx context.Context, schoolUrl, sessionID, userEventID, participatorID, supportID string) error {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	if err := s.sessionService.SetSessionLanguage(ctx, schoolUrl, sessionID); err != nil {
		return fmt.Errorf("failed to set session language: %w", err)
	}

	endpoint := fmt.Sprintf("%s/ajax/ajax_aktivitetsanmalan.jsp", strings.TrimSuffix(schoolUrl, "/"))
	params := map[string]string{
		"op":                     "tabortStod",
		"aktivitetsTillfallesId": userEventID,
		"stodId":                 supportID,
		"deltagarId":             participatorID,
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to remove event support: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to remove event support, status code: %d", response.StatusCode)
	}

	return nil
}
