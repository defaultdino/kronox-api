package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
)

type EventHandler struct {
	eventService *services.EventService
}

func NewEventHandler(eventService *services.EventService, parserService parsers.ParserService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

func (h *EventHandler) GetUserEvents(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	events, err := AttemptOverSchoolURLs(c, func(url string) (*parsers.EventsResponse, error) {
		return h.eventService.GetUserEvents(c.Request.Context(), url, sessionID)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

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
