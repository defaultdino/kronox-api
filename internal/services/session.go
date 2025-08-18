// internal/services/session_service.go
package services

import (
	"context"
	"fmt"
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

func (s *SessionService) AuthenticateAndGetSession(ctx context.Context, school, username, password string) (string, error) {
	return "session_token", nil
}

func (s *SessionService) ValidateSession(ctx context.Context, school, sessionID string) bool {
	return true
}

func (s *SessionService) SetSessionLanguage(ctx context.Context, school, sessionID string) error {
	type contextKey string
	const sessionIDKey contextKey = "session_id"
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/ajax/ajax_setSessionLanguage.jsp", strings.TrimSuffix(school, "/"))
	params := map[string]string{
		"lang": "en",
	}

	_, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	return err
}
