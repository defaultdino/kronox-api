package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	SessionID string             `json:"session_id"`
	ExpiresAt time.Time          `json:"expires_at"`
	UserInfo  *services.UserInfo `json:"user_info"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sessionID string
	var userInfo *services.UserInfo
	var loginErr error

	_, _ = AttemptOverSchoolURLs(c, func(url string) (bool, error) {
		sessionID, userInfo, loginErr = h.authService.Login(c.Request.Context(), req.Username, req.Password, url)
		return loginErr == nil, loginErr
	})

	if loginErr != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials or login failed"})
		return
	}

	response := LoginResponse{
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserInfo:  userInfo,
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) ValidateSession(c *gin.Context) {
	var sessionID string

	if auth := c.GetHeader("Authorization"); auth != "" {
		if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
			sessionID = after
		}
	}

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id and school required"})
		return
	}

	var isValid bool
	var err error

	_, _ = AttemptOverSchoolURLs(c, func(url string) (bool, error) {
		isValid, err = h.authService.ValidateSession(c.Request.Context(), sessionID, url)
		return err == nil, err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": isValid})
}
