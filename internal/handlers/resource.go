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

// GetAllResources godoc
// @Summary      Get all resources
// @Description  Retrieve all available resources for booking across multiple school URLs
// @Tags         resources
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  ResourcesListResponse  "List of available resources"
// @Failure      401           {object}  ErrorResponse          "Session required"
// @Failure      500           {object}  ErrorResponse          "Internal server error"
// @Security     BearerAuth
// @Router       /resources [get]
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

// GetBookings godoc
// @Summary      Get user bookings
// @Description  Retrieve all bookings made by the authenticated user
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  BookingsListResponse  "List of user bookings"
// @Failure      401           {object}  ErrorResponse         "Session required"
// @Failure      500           {object}  ErrorResponse         "Internal server error"
// @Security     BearerAuth
// @Router       /resources/bookings [get]
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

// GetActiveBookingsForResource godoc
// @Summary      Get active bookings for resource
// @Description  Retrieve all active bookings for a specific resource
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        resourceId     path      string  true  "Resource ID to get bookings for"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  BookingsListResponse  "List of active bookings for the resource"
// @Failure      400           {object}  ErrorResponse         "resourceId path parameter required"
// @Failure      401           {object}  ErrorResponse         "Session required"
// @Failure      500           {object}  ErrorResponse         "Internal server error"
// @Security     BearerAuth
// @Router       /resources/{resourceId}/bookings [get]
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

// GetResourceAvailability godoc
// @Summary      Get resource availability
// @Description  Get available time slots for a specific resource on a given date
// @Tags         resources
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        resource_id    query     string  true  "Resource ID to check availability for"
// @Param        date          query     string  true  "Date in YYYY-MM-DD format"  Format(date)
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  AvailabilityResponse  "Available time slots for the resource"
// @Failure      400           {object}  ErrorResponse         "resource_id and date are required"
// @Failure      401           {object}  ErrorResponse         "Session required"
// @Failure      500           {object}  ErrorResponse         "Internal server error"
// @Security     BearerAuth
// @Router       /resources/availability [get]
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

// BookResource godoc
// @Summary      Book a resource
// @Description  Create a new booking for a specific resource and time slot
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        Authorization  header                    string                   true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        resourceId     path                      string                   true  "Resource ID to book"
// @Param        booking        body                      booking.BookingRequest   true  "Booking details"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}                  SuccessResponse          "Booking successful"
// @Failure      400           {object}                  ErrorResponse            "Invalid request data"
// @Failure      401           {object}                  ErrorResponse            "Session required"
// @Failure      500           {object}                  ErrorResponse            "Internal server error"
// @Security     BearerAuth
// @Router       /resources/{resourceId}/book [post]
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

// UnbookResource godoc
// @Summary      Cancel a booking
// @Description  Cancel an existing resource booking
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        bookingId      path      string  true  "Booking ID to cancel"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  SuccessResponse  "Booking cancelled successfully"
// @Failure      400           {object}  ErrorResponse    "bookingId is required"
// @Failure      401           {object}  ErrorResponse    "Session required"
// @Failure      500           {object}  ErrorResponse    "Internal server error"
// @Security     BearerAuth
// @Router       /resources/bookings/{bookingId} [delete]
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

// UnbookResourceByID godoc
// @Summary      Cancel booking by ID (alternative endpoint)
// @Description  Cancel an existing resource booking using booking ID
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        bookingId      path      string  true  "Booking ID to cancel"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  SuccessResponse  "Booking cancelled successfully"
// @Failure      400           {object}  ErrorResponse    "bookingId path parameter required"
// @Failure      401           {object}  ErrorResponse    "Session required"
// @Failure      500           {object}  ErrorResponse    "Internal server error"
// @Security     BearerAuth
// @Router       /bookings/{bookingId}/cancel [delete]
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

// ConfirmResourceBooking godoc
// @Summary      Confirm a booking
// @Description  Confirm an existing resource booking with resource ID in request body
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        Authorization  header                         string                          true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        bookingId      path                           string                          true  "Booking ID to confirm"
// @Param        confirmation   body                           booking.ConfirmBookingRequest   true  "Confirmation details"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}                       SuccessResponse                 "Booking confirmed successfully"
// @Failure      400           {object}                       ErrorResponse                   "Invalid request data"
// @Failure      401           {object}                       ErrorResponse                   "Session required"
// @Failure      500           {object}                       ErrorResponse                   "Internal server error"
// @Security     BearerAuth
// @Router       /resources/bookings/{bookingId}/confirm [post]
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

// ConfirmResourceBookingByID godoc
// @Summary      Confirm booking by ID (path parameters)
// @Description  Confirm an existing resource booking using resource ID and booking ID from path
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        bookingId      path      string  true  "Booking ID to confirm"
// @Param        resourceId     path      string  true  "Resource ID for confirmation"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  SuccessResponse  "Booking confirmed successfully"
// @Failure      400           {object}  ErrorResponse    "bookingId or resourceId path parameter required"
// @Failure      401           {object}  ErrorResponse    "Session required"
// @Failure      500           {object}  ErrorResponse    "Internal server error"
// @Security     BearerAuth
// @Router       /bookings/{bookingId}/confirm/{resourceId} [post]
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

// GetAvailableResources godoc
// @Summary      Get available resources (alternative endpoint)
// @Description  Get available time slots for a specific resource on a given date using path parameter
// @Tags         resources
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string  true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        resourceId     path      string  true  "Resource ID to check availability for"
// @Param        date          query     string  true  "Date in YYYY-MM-DD format"  Format(date)
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}  AvailabilityResponse  "Available time slots for the resource"
// @Failure      400           {object}  ErrorResponse         "resourceId and date are required"
// @Failure      401           {object}  ErrorResponse         "Session required"
// @Failure      500           {object}  ErrorResponse         "Internal server error"
// @Security     BearerAuth
// @Router       /resources/{resourceId}/availability [get]
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

// ConfirmResourceBookingWithBody godoc
// @Summary      Confirm booking with resource ID in body
// @Description  Confirm an existing resource booking with resource ID provided in request body
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        Authorization  header                string                     true  "Bearer token (session ID)"  Format(Bearer {session_id})
// @Param        bookingId      path                  string                     true  "Booking ID to confirm"
// @Param        resource       body                  ConfirmBookingBodyRequest  true  "Resource ID for confirmation"
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Success      200           {object}              SuccessResponse            "Booking confirmed successfully"
// @Failure      400           {object}              ErrorResponse              "bookingId path parameter or resourceId in body required"
// @Failure      401           {object}              ErrorResponse              "Session required"
// @Failure      500           {object}              ErrorResponse              "Internal server error"
// @Security     BearerAuth
// @Router       /bookings/{bookingId}/confirm [post]
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

// ResourcesListResponse represents the response for getting resources
// @Description Response containing list of available resources
type ResourcesListResponse struct {
	Resources []*booking.Resource `json:"resources"`
}

// BookingsListResponse represents the response for getting bookings
// @Description Response containing list of bookings
type BookingsListResponse struct {
	Bookings []*booking.Booking `json:"bookings"`
}

// AvailabilityResponse represents the response for resource availability
// @Description Response containing available time slots
type AvailabilityResponse struct {
	Availability []*booking.AvailabilitySlot `json:"availability"`
}

// ConfirmBookingBodyRequest represents request body for booking confirmation
// @Description Request body for confirming a booking with resource ID
type ConfirmBookingBodyRequest struct {
	ResourceID string `json:"resourceId" binding:"required" example:"res_123"`
}
