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
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	"github.com/tumble-for-kronox/kronox-api/pkg/models/user"
)

type AuthService struct {
	app           *app.App
	parserService parsers.ParserService
}

func NewAuthService(app *app.App, parserService parsers.ParserService) *AuthService {
	return &AuthService{
		app:           app,
		parserService: parserService,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password, schoolUrl string) (*user.User, error) {
	if err := s.app.KronoxClient.ResetCookieJar(); err != nil {
		log.Printf("Warning: Failed to reset cookie jar: %v\n", err)
	}

	endpoint := fmt.Sprintf("%s/login_do.jsp", strings.TrimSuffix(schoolUrl, "/"))

	postData := fmt.Sprintf("username=%s&password=%s",
		url.QueryEscape(username),
		url.QueryEscape(password))

	resp, err := s.app.KronoxClient.SendRequestWithBody(ctx, http.MethodPost, endpoint, map[string]string{}, postData)
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

	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if location == "" {
			return nil, fmt.Errorf("redirect response missing Location header")
		}

		ctxWithSession := context.WithValue(ctx, sessionIDKey, sessionID)

		redirectResp, err := s.app.KronoxClient.SendRequest(ctxWithSession, http.MethodGet, location, map[string]string{})
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
			Name:     userInfo.Name,
			Username: userInfo.Username,
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
		Name:     userInfo.Name,
		Username: userInfo.Username,
	}, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, sessionID, schoolUrl string) (bool, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/start.jsp", strings.TrimSuffix(schoolUrl, "/"))
	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, map[string]string{})
	if err != nil {
		return false, fmt.Errorf("failed to validate session: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return false, fmt.Errorf("failed to read validation response: %w", err)
		}

		bodyStr := string(body)

		// when authenticated: shows "Hej [Name]" and navigation links
		// when not authenticated: shows login form with "Användarnamn:" and "Lösenord:"
		isAuthenticated := strings.Contains(bodyStr, "Hej ") &&
			!strings.Contains(bodyStr, "Användarnamn:") &&
			!strings.Contains(bodyStr, "Lösenord:")

		return isAuthenticated, nil
	}

	return false, nil
}
