package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
	"github.com/tumble-for-kronox/kronox-api/pkg/models/booking"
)

type BookingHandler struct {
	bookingService *services.BookingService
}

func NewBookingHandler(bookingService *services.BookingService) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
	}
}

func (h *BookingHandler) GetUserBookings(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	school := c.Query("school")
	if school == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "school required"})
		return
	}

	bookings, err := h.bookingService.GetUserBookings(c.Request.Context(), school, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func (h *BookingHandler) GetResourceAvailability(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	school := c.Query("school")
	if school == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "school required"})
		return
	}

	resourceID := c.Query("resource_id")
	if resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource_id required"})
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date required (YYYY-MM-DD format)"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	availability, err := h.bookingService.GetResourceAvailability(c.Request.Context(), school, sessionID, date, resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"availability": availability})
}

func (h *BookingHandler) BookResource(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	school := c.Query("school")
	if school == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "school required"})
		return
	}

	var req booking.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.bookingService.BookResource(c.Request.Context(), school, sessionID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking successful"})
}
