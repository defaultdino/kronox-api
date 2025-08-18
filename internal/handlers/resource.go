package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
	booking "github.com/tumble-for-kronox/kronox-api/pkg/models/resource"
)

type ResourceHandler struct {
	resourceService *services.ResourceService
}

func NewResourceHandler(resourceService *services.ResourceService) *ResourceHandler {
	return &ResourceHandler{
		resourceService: resourceService,
	}
}

func (h *ResourceHandler) GetAllResources(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	resources, err := AttemptOverSchoolURLs(c, func(url string) ([]*booking.Resource, error) {
		return h.resourceService.GetResources(c.Request.Context(), url, sessionID)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resources": resources})
}

func (h *ResourceHandler) GetBookings(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	bookings, err := AttemptOverSchoolURLs(c, func(url string) ([]*booking.Booking, error) {
		return h.resourceService.GetBookedResources(c.Request.Context(), url, sessionID)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func (h *ResourceHandler) GetActiveBookingsForResource(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	resourceID := c.Param("resourceId")
	if resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceId path parameter required"})
		return
	}

	bookings, err := AttemptOverSchoolURLs(c, func(url string) ([]*booking.Booking, error) {
		return h.resourceService.GetActiveResourceBookings(c.Request.Context(), url, sessionID, resourceID)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func (h *ResourceHandler) GetResourceAvailability(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
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

	availability, err := AttemptOverSchoolURLs(c, func(url string) ([]*booking.AvailabilitySlot, error) {
		return h.resourceService.GetAvailableResources(c.Request.Context(), url, sessionID, date, resourceID)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"availability": availability})
}

func (h *ResourceHandler) BookResource(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	resourceId := c.Param("resourceId")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	var req booking.BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if resourceId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceId is required"})
		return
	}

	if req.Slot == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slot is required"})
		return
	}

	if req.Slot.LocationId == nil || *req.Slot.LocationId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slot.locationId is required"})
		return
	}

	if req.Slot.ResourceType == nil || *req.Slot.ResourceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slot.resourceType is required"})
		return
	}

	if req.Slot.TimeSlotId == nil || *req.Slot.TimeSlotId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slot.timeSlotId is required"})
		return
	}

	if err := AttemptOverSchoolURLsBool(c, func(url string) error {
		return h.resourceService.BookResource(c.Request.Context(), url, sessionID, &req, resourceId)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking successful"})
}

func (h *ResourceHandler) UnbookResource(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	bookingId := c.Param("bookingId")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	if bookingId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bookingId is required"})
		return
	}

	if err := AttemptOverSchoolURLsBool(c, func(url string) error {
		return h.resourceService.UnbookResource(c.Request.Context(), url, sessionID, bookingId)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled successfully"})
}

func (h *ResourceHandler) UnbookResourceByID(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	bookingID := c.Param("bookingId")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bookingId path parameter required"})
		return
	}

	if err := AttemptOverSchoolURLsBool(c, func(url string) error {
		return h.resourceService.UnbookResource(c.Request.Context(), url, sessionID, bookingID)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking cancelled successfully"})
}

func (h *ResourceHandler) ConfirmResourceBooking(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	bookingId := c.Param("bookingId")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	var req booking.ConfirmBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.ResourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceId is required"})
		return
	}

	if bookingId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bookingId is required"})
		return
	}

	if err := AttemptOverSchoolURLsBool(c, func(url string) error {
		return h.resourceService.ConfirmBooking(c.Request.Context(), url, sessionID, bookingId, req.ResourceID)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking confirmed successfully"})
}

func (h *ResourceHandler) ConfirmResourceBookingByID(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	bookingID := c.Param("bookingId")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bookingId path parameter required"})
		return
	}

	resourceID := c.Param("resourceId")
	if resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceId path parameter required"})
		return
	}

	if err := AttemptOverSchoolURLsBool(c, func(url string) error {
		return h.resourceService.ConfirmBooking(c.Request.Context(), url, sessionID, bookingID, resourceID)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking confirmed successfully"})
}

func (h *ResourceHandler) GetAvailableResources(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	resourceID := c.Param("resourceId")
	if resourceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceId path parameter required"})
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

	availability, err := AttemptOverSchoolURLs(c, func(url string) ([]*booking.AvailabilitySlot, error) {
		return h.resourceService.GetAvailableResources(c.Request.Context(), url, sessionID, date, resourceID)
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"availability": availability})
}

func (h *ResourceHandler) ConfirmResourceBookingWithBody(c *gin.Context) {
	sessionID, exists := middleware.GetSessionID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "session required"})
		return
	}

	bookingID := c.Param("bookingId")
	if bookingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bookingId path parameter required"})
		return
	}

	var req struct {
		ResourceID string `json:"resourceId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceId is required in request body"})
		return
	}

	if err := AttemptOverSchoolURLsBool(c, func(url string) error {
		return h.resourceService.ConfirmBooking(c.Request.Context(), url, sessionID, bookingID, req.ResourceID)
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking confirmed successfully"})
}
