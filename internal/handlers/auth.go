package handlers

import (
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
// @Description  Authenticate user with username and password across multiple school URLs
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Param        credentials  body      user.LoginRequest  true  "Login credentials"
// @Success      200         {object}  user.User          "User successfully authenticated"
// @Failure      400         {object}  ErrorResponse      "Invalid request body"
// @Failure      401         {object}  ErrorResponse      "Invalid credentials or login failed"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// "cache" for getting events & bookings later
	var lastSchoolUrl string

	user, err := AttemptOverSchoolURLs(c, func(url string) (*user.User, error) {
		lastSchoolUrl = url
		return h.authService.Login(c.Request.Context(), req.Username, req.Password, url)
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials or login failed"})
		return
	}

	registeredEvents, errRegisteredEvents := h.eventService.GetUserEvents(c, lastSchoolUrl, user.SessionID)
	bookedResources, errBookedResources := h.resourceService.GetBookedResources(c, lastSchoolUrl, user.SessionID)

	if errRegisteredEvents != nil || errBookedResources != nil {
		// we got the user just not their events OR resources
		// for caching in mongo or returning in the object,
		// thus they could be null in tumble-api
		c.JSON(http.StatusOK, user)
	}

	user.Events = registeredEvents.Registered
	user.Bookings = bookedResources

	c.JSON(http.StatusOK, user)
}

// ValidateSession godoc
// @Summary      Validate user session
// @Description  Check if a user session is valid across multiple school URLs
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Success      200           {object}  SessionValidationResponse  "Session validation result"
// @Failure      400           {object}  ErrorResponse              "Missing session_id in Authorization header"
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

// SessionValidationResponse represents session validation response
// @Description Session validation response structure
type SessionValidationResponse struct {
	Valid bool `json:"valid" example:"true"`
}
