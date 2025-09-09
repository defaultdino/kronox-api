package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/tumble-for-kronox/kronox-api/internal/app"
)

type SessionService struct {
	app *app.App
}

func NewSessionService(app *app.App) *SessionService {
	return &SessionService{app: app}
}

func (s *SessionService) ValidateSession(ctx context.Context, school, sessionID string) bool {
	isValid, _ := s.RefreshSession(ctx, school, sessionID)
	if isValid {
		log.Printf("Session validation via poll: VALID\n")
		return true
	}

	ctx = context.WithValue(ctx, sessionIDKey, sessionID)
	endpoint := fmt.Sprintf("%s/start.jsp", strings.TrimSuffix(school, "/"))

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, map[string]string{})
	if err != nil {
		return false
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false
	}

	bodyStr := string(body)

	isAuthenticated := strings.Contains(bodyStr, "Hej ") &&
		!strings.Contains(bodyStr, "Användarnamn:") &&
		!strings.Contains(bodyStr, "Lösenord:")

	return isAuthenticated
}

func (s *SessionService) SetSessionLanguage(ctx context.Context, schoolUrl, sessionID string) error {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/ajax/ajax_lang.jsp", strings.TrimSuffix(schoolUrl, "/"))
	params := map[string]string{
		"lang": "EN",
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to set session language: %w", err)
	}
	defer response.Body.Close()

	return nil
}

func (s *SessionService) RefreshSession(ctx context.Context, schoolUrl, sessionID string) (bool, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/ajax/ajax_session.jsp", strings.TrimSuffix(schoolUrl, "/"))
	params := map[string]string{
		"op": "poll",
	}

	response, err := s.app.KronoxClient.SendRequestWithBody(ctx, http.MethodPost, endpoint, params, "")
	if err != nil {
		return false, fmt.Errorf("failed to refresh session: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read refresh response: %w", err)
	}

	content := strings.TrimSpace(string(body))
	isOK := strings.EqualFold(content, "OK")

	return isOK, nil
}
