package handlers

import (
	"net/http"
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
	School   string `json:"school" binding:"required"`
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

	sessionID, userInfo, err := h.authService.Login(c.Request.Context(), req.Username, req.Password, req.School)
	if err != nil {
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
	sessionID := c.GetHeader("X-Session-ID")
	school := c.Query("school")

	if sessionID == "" || school == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id and school required"})
		return
	}

	isValid, err := h.authService.ValidateSession(c.Request.Context(), sessionID, school)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": isValid})
}
