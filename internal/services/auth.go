package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tumble-for-kronox/kronox-api/internal/app"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
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

type UserInfo struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

func (s *AuthService) Login(ctx context.Context, username, password, school string) (string, *UserInfo, error) {
	endpoint := fmt.Sprintf("%s/login_do.jsp", strings.TrimSuffix(school, "/"))

	postData := fmt.Sprintf("username=%s&password=%s",
		url.QueryEscape(username),
		url.QueryEscape(password))

	resp, err := s.app.KronoxClient.SendRequestWithBody(ctx, http.MethodPost, endpoint, map[string]string{}, postData)
	if err != nil {
		return "", nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read response: %w", err)
	}
	responseHTML := string(body)

	// we can't access the cookie jar directly through the interface
	// so we need to check if cookies were sent back via response headers
	var sessionID string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "JSESSIONID" {
			sessionID = cookie.Value
			break
		}
	}

	if sessionID == "" {
		return "", nil, fmt.Errorf("no session cookie found - login likely failed")
	}

	userInfo, err := s.parserService.ParseUserLogin(responseHTML)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return sessionID, &UserInfo{
		Name:     userInfo.Name,
		Username: userInfo.Username,
	}, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, sessionID, school string) (bool, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/dashboard.jsp", strings.TrimSuffix(school, "/"))
	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, map[string]string{})
	if err != nil {
		return false, nil
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return !strings.Contains(string(body), "login"), nil
	}

	return false, nil
}
