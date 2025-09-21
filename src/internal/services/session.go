package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/app"
	"github.com/tumble-for-kronox/kronox-api/internal/clients/kronox"
)

type UserSession struct {
	Client     kronox.Client
	SessionID  string
	Username   string
	SchoolURL  string
	CreatedAt  time.Time
	LastUsedAt time.Time
}

type SessionManager struct {
	sessions map[string]*UserSession
	mutex    sync.RWMutex
	app      *app.App
}

func NewSessionManager(app *app.App) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*UserSession),
		app:      app,
	}

	go sm.cleanupExpiredSessions()
	return sm
}

func (sm *SessionManager) CreateSession(sessionID, username, schoolURL string) *UserSession {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	client := kronox.NewClient(httpClient)

	session := &UserSession{
		Client:     client,
		SessionID:  sessionID,
		Username:   username,
		SchoolURL:  schoolURL,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}

	sm.sessions[sessionID] = session
	return session
}

func (sm *SessionManager) GetSession(sessionID string) (*UserSession, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if exists {
		session.LastUsedAt = time.Now()
	}

	return session, exists
}

func (sm *SessionManager) RemoveSession(sessionID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.sessions, sessionID)
}

func (sm *SessionManager) GetSessionCount() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return len(sm.sessions)
}

func (sm *SessionManager) ValidateAndCleanupSession(htmlContent string, statusCode int, sessionID string) error {
	htmlLower := strings.ToLower(htmlContent)

	if strings.Contains(htmlLower, `<form id="loginform">`) {
		fmt.Fprintf(gin.DefaultWriter, "Session %s invalid: Login form detected, removing from manager\n", sessionID)
		sm.RemoveSession(sessionID)
		return fmt.Errorf("session expired - redirected to login page")
	}
	return nil
}

func (sm *SessionManager) ValidateSession(ctx context.Context, sessionID, schoolURL string) (bool, error) {
	userSession, exists := sm.GetSession(sessionID)
	if !exists {
		return false, fmt.Errorf("session not found")
	}

	endpoint := fmt.Sprintf("%s/start.jsp", strings.TrimSuffix(schoolURL, "/"))
	ctxWithSession := context.WithValue(ctx, sessionIDKey, sessionID)

	response, err := userSession.Client.SendRequest(ctxWithSession, http.MethodGet, endpoint, map[string]string{})
	if err != nil {
		return false, fmt.Errorf("failed to validate session: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return false, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read validation response: %w", err)
	}

	bodyStr := string(body)
	isAuthenticated := strings.Contains(bodyStr, "Hej ") &&
		!strings.Contains(bodyStr, "Användarnamn:") &&
		!strings.Contains(bodyStr, "Lösenord:")

	if !isAuthenticated {
		sm.RemoveSession(sessionID)
	}

	return isAuthenticated, nil
}

func (sm *SessionManager) SetSessionLanguage(ctx context.Context, sessionID, schoolURL string) error {
	userSession, exists := sm.GetSession(sessionID)
	if !exists {
		return fmt.Errorf("session not found")
	}

	endpoint := fmt.Sprintf("%s/ajax/ajax_lang.jsp", strings.TrimSuffix(schoolURL, "/"))
	params := map[string]string{
		"lang": "EN",
	}

	ctxWithSession := context.WithValue(ctx, sessionIDKey, sessionID)
	response, err := userSession.Client.SendRequest(ctxWithSession, http.MethodGet, endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to set session language: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set session language, status code: %d", response.StatusCode)
	}

	return nil
}

func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mutex.Lock()
		now := time.Now()

		for sessionID, session := range sm.sessions {
			if now.Sub(session.LastUsedAt) > 2*time.Hour {
				delete(sm.sessions, sessionID)
			}
		}

		sm.mutex.Unlock()
	}
}
