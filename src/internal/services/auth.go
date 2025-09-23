package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/tumble-for-kronox/kronox-api/internal/app"
	"github.com/tumble-for-kronox/kronox-api/internal/clients/kronox"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	"github.com/tumble-for-kronox/kronox-api/pkg/models/user"
)

type AuthService struct {
	app            *app.App
	parserService  parsers.ParserService
	sessionManager *SessionManager
}

func NewAuthService(app *app.App, parserService parsers.ParserService, sessionManager *SessionManager) *AuthService {
	return &AuthService{
		app:            app,
		parserService:  parserService,
		sessionManager: sessionManager,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password, schoolUrl string) (*user.User, error) {
	client := s.app.NewKronoxClient()

	endpoint := fmt.Sprintf("%s/login_do.jsp", strings.TrimSuffix(schoolUrl, "/"))

	postData := fmt.Sprintf("username=%s&password=%s",
		url.QueryEscape(username),
		url.QueryEscape(password))

	resp, err := client.SendRequestWithFormData(ctx, http.MethodPost, endpoint, map[string]string{}, postData)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	var sessionID string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			sessionID = cookie.Value
			break
		}
	}

	if sessionID == "" {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Login failed - no session cookie. Response body: %s\n", string(body))
		return nil, fmt.Errorf("no session cookie found - login likely failed")
	}

	userSession := s.sessionManager.CreateSession(sessionID, username, schoolUrl)

	if err := s.copyAuthenticationState(client, userSession.Client, schoolUrl); err != nil {
		return nil, fmt.Errorf("failed to copy authentication state: %w", err)
	}

	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if location == "" {
			return nil, fmt.Errorf("redirect response missing Location header")
		}

		ctxWithSession := context.WithValue(ctx, sessionIDKey, sessionID)
		redirectResp, err := userSession.Client.SendRequest(ctxWithSession, http.MethodGet, location, map[string]string{})
		if err != nil {
			return nil, fmt.Errorf("failed to follow redirect: %w", err)
		}
		defer redirectResp.Body.Close()

		body, err := io.ReadAll(redirectResp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read redirect response: %w", err)
		}
		responseHTML := string(body)

		userInfo, err := s.parserService.ParseUserLogin(responseHTML)
		if err != nil {
			return nil, fmt.Errorf("failed to parse user info: %w", err)
		}

		return &user.User{
			Name:      userInfo.Name,
			Username:  userInfo.Username,
			SessionID: sessionID,
		}, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	responseHTML := string(body)

	if strings.Contains(strings.ToLower(responseHTML), "error") ||
		strings.Contains(strings.ToLower(responseHTML), "invalid") ||
		strings.Contains(strings.ToLower(responseHTML), "fel") {
		return nil, fmt.Errorf("login failed - error detected in response")
	}

	userInfo, err := s.parserService.ParseUserLogin(responseHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &user.User{
		Name:      userInfo.Name,
		Username:  userInfo.Username,
		SessionID: sessionID,
	}, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, sessionID, schoolUrl string) (bool, error) {
	return s.sessionManager.ValidateSession(ctx, sessionID, schoolUrl)
}

func (s *AuthService) Logout(sessionID string) error {
	s.sessionManager.RemoveSession(sessionID)
	return nil
}

func (s *AuthService) copyAuthenticationState(loginClient, sessionClient kronox.Client, schoolUrl string) error {
	return sessionClient.CopyCookiesFrom(loginClient, schoolUrl)
}

func (s *AuthService) PollSession(ctx context.Context, sessionID, schoolUrl string) (string, error) {
	userSession, exists := s.sessionManager.GetSession(sessionID)
	if !exists {
		return "SESSIONSFEL", fmt.Errorf("session not found")
	}

	endpoint := fmt.Sprintf("%s/ajax/ajax_session.jsp", strings.TrimSuffix(schoolUrl, "/"))

	ctxWithSession := context.WithValue(ctx, sessionIDKey, sessionID)
	resp, err := userSession.Client.SendRequest(ctxWithSession, http.MethodGet, endpoint, map[string]string{
		"op": "poll",
	})
	if err != nil {
		return "ERROR", fmt.Errorf("poll request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "ERROR", fmt.Errorf("failed to read response: %w", err)
	}

	response := strings.TrimSpace(string(body))

	switch response {
	case "OK":
		return "OK", nil
	case "SESSIONSFEL", "ERROR":
		return response, fmt.Errorf("session invalid: %s", response)
	}

	return "ERROR", fmt.Errorf("unexpected response: %s", response)
}
