package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
	"github.com/tumble-for-kronox/kronox-api/pkg/models/user"
)

type AuthHandler struct {
	authService     *services.AuthService
	eventService    *services.EventService
	resourceService *services.ResourceService
}

func NewAuthHandler(
	authService *services.AuthService,
	eventService *services.EventService,
	resourceService *services.ResourceService,
) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		eventService:    eventService,
		resourceService: resourceService,
	}
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with username and password using specified school URL
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Param        url_index query     int     true  "Index of the school URL to use"  example(0)
// @Param        credentials  body      user.LoginRequest  true  "Login credentials"
// @Success      200         {object}  user.User          "User successfully authenticated"
// @Failure      400         {object}  ErrorResponse      "Invalid request body or url_index"
// @Failure      401         {object}  ErrorResponse      "Invalid credentials or login failed"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schoolURL, err := GetSchoolURLFromIndex(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	user, err := h.authService.Login(ctx, req.Username, req.Password, schoolURL)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials or login failed"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Add this to your existing AuthHandler in kronox-api

// PollSession godoc
// @Summary      Poll user session
// @Description  Poll KronoX session to keep it alive (mimics web frontend behavior)
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Param        url_index query     int     true  "Index of the school URL to use"  example(0)
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Success      200           {object}  SessionPollResponse    "Session poll result"
// @Failure      400           {object}  ErrorResponse          "Missing session_id in Authorization header or invalid url_index"
// @Failure      500           {object}  ErrorResponse          "Session poll failed"
// @Security     BearerAuth
// @Router       /auth/poll [get]
func (h *AuthHandler) PollSession(c *gin.Context) {
	var sessionID string

	if auth := c.GetHeader("Authorization"); auth != "" {
		if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
			sessionID = after
		}
	}

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id required in Authorization header"})
		return
	}

	schoolURL, err := GetSchoolURLFromIndex(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	status, err := h.authService.PollSession(ctx, sessionID, schoolURL)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"status": status,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// SessionPollResponse represents session poll response
// @Description Session poll response structure
type SessionPollResponse struct {
	Status string `json:"status" example:"OK"`
}

// ValidateSession godoc
// @Summary      Validate user session
// @Description  Check if a user session is valid using specified school URL
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Param        url_index query     int     true  "Index of the school URL to use"  example(0)
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Success      200           {object}  SessionValidationResponse  "Session validation result"
// @Failure      400           {object}  ErrorResponse              "Missing session_id in Authorization header or invalid url_index"
// @Failure      500           {object}  ErrorResponse              "Internal server error during validation"
// @Security     BearerAuth
// @Router       /auth/validate [get]
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

	schoolURL, err := GetSchoolURLFromIndex(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	isValid, err := h.authService.ValidateSession(ctx, sessionID, schoolURL)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": isValid})
}

// SessionValidationResponse represents session validation response
// @Description Session validation response structure
type SessionValidationResponse struct {
	Valid bool `json:"valid" example:"true"`
}
