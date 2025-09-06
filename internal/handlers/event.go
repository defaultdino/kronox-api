package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
	"github.com/tumble-for-kronox/kronox-api/pkg/middleware"
	"github.com/tumble-for-kronox/kronox-api/pkg/models/user"
)

type EventHandler struct {
	eventService *services.EventService
}

func NewEventHandler(eventService *services.EventService, parserService parsers.ParserService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

// GetUserEvents godoc
// @Summary      Get user events
// @Description  Retrieve all events for the authenticated user across multiple school URLs
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  EventsListResponse  "List of user events"
// @Failure      401           {object}  ErrorResponse       "Session required"
// @Failure      500           {object}  ErrorResponse       "Internal server error"
// @Security     BearerAuth
// @Router       /events [get]
func (h *EventHandler) GetUserEvents(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	events, err := AttemptOverSchoolURLs(c, func(url string) (*user.EventsResponse, error) {
		return h.eventService.GetUserEvents(c.Request.Context(), url, sessionID)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

// RegisterUserEvent godoc
// @Summary      Register for event
// @Description  Register the authenticated user for a specific event
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        eventId        path      string  true  "Event ID to register for"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  SuccessResponse  "Successfully registered for event"
// @Failure      400           {object}  ErrorResponse    "eventId path parameter required"
// @Failure      401           {object}  ErrorResponse    "Session required"
// @Failure      500           {object}  ErrorResponse    "Internal server error"
// @Security     BearerAuth
// @Router       /events/{eventId}/register [post]
func (h *EventHandler) RegisterUserEvent(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	userEventID := c.Param("eventId")
	if userEventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eventId path parameter required"})
		return
	}

	if err := AttemptOverSchoolURLsBool(c, func(url string) error {
		return h.eventService.RegisterUserEvent(c.Request.Context(), url, sessionID, userEventID)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully registered for event"})
}

// UnregisterUserEvent godoc
// @Summary      Unregister from event
// @Description  Unregister the authenticated user from a specific event
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        eventId        path      string  true  "Event ID to unregister from"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  SuccessResponse  "Successfully unregistered from event"
// @Failure      400           {object}  ErrorResponse    "eventId path parameter required"
// @Failure      401           {object}  ErrorResponse    "Session required"
// @Failure      500           {object}  ErrorResponse    "Internal server error"
// @Security     BearerAuth
// @Router       /events/{eventId}/unregister [delete]
func (h *EventHandler) UnregisterUserEvent(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	userEventID := c.Param("eventId")
	if userEventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eventId path parameter required"})
		return
	}

	if err := AttemptOverSchoolURLsBool(c, func(url string) error {
		return h.eventService.UnregisterUserEvent(c.Request.Context(), url, sessionID, userEventID)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully unregistered from event"})
}

// AddEventSupport godoc
// @Summary      Add event support
// @Description  Add support for a specific participator in an event
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        Authorization    header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        participatorId   path      string  true  "Participator ID to support"
// @Param        supportId        path      string  true  "Support ID to add"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200             {object}  SuccessResponse  "Successfully added event support"
// @Failure      400             {object}  ErrorResponse    "participatorId or supportId path parameter required"
// @Failure      401             {object}  ErrorResponse    "Session required"
// @Failure      500             {object}  ErrorResponse    "Internal server error"
// @Security     BearerAuth
// @Router       /events/support/{participatorId}/{supportId} [post]
func (h *EventHandler) AddEventSupport(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	participatorID := c.Param("participatorId")
	if participatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "participatorId path parameter required"})
		return
	}

	supportID := c.Param("supportId")
	if supportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "supportId path parameter required"})
		return
	}

	if err := AttemptOverSchoolURLsBool(c, func(url string) error {
		return h.eventService.AddEventSupport(c.Request.Context(), url, sessionID, participatorID, supportID)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully added event support"})
}

// RemoveEventSupport godoc
// @Summary      Remove event support
// @Description  Remove support for a specific participator in an event
// @Tags         events
// @Accept       json
// @Produce      json
// @Param        Authorization    header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        eventId          path      string  true  "Event ID"
// @Param        participatorId   path      string  true  "Participator ID"
// @Param        supportId        path      string  true  "Support ID to remove"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200             {object}  SuccessResponse  "Successfully removed event support"
// @Failure      400             {object}  ErrorResponse    "eventId, participatorId, or supportId path parameter required"
// @Failure      401             {object}  ErrorResponse    "Session required"
// @Failure      500             {object}  ErrorResponse    "Internal server error"
// @Security     BearerAuth
// @Router       /events/{eventId}/support/{participatorId}/{supportId} [delete]
func (h *EventHandler) RemoveEventSupport(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	userEventID := c.Param("eventId")
	if userEventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eventId path parameter required"})
		return
	}

	participatorID := c.Param("participatorId")
	if participatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "participatorId path parameter required"})
		return
	}

	supportID := c.Param("supportId")
	if supportID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "supportId path parameter required"})
		return
	}

	if err := AttemptOverSchoolURLsBool(c, func(url string) error {
		return h.eventService.RemoveEventSupport(c.Request.Context(), url, sessionID, userEventID, participatorID, supportID)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully removed event support"})
}

// EventsListResponse represents the response for getting user events
// @Description Response containing list of user events
type EventsListResponse struct {
	Events *user.EventsResponse `json:"events"`
}

// SuccessResponse represents a successful operation response
// @Description Standard success response
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// ErrorResponse represents an error response
// @Description Error response structure
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request"`
}
