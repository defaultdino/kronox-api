package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/tumble-for-kronox/kronox-api/internal/app"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	booking "github.com/tumble-for-kronox/kronox-api/pkg/models/resource"
)

type ResourceService struct {
	app            *app.App
	sessionService *SessionService
	parserService  parsers.ParserService
}

func NewResourceService(app *app.App, sessionService *SessionService, parserService parsers.ParserService) *ResourceService {
	return &ResourceService{
		app:            app,
		sessionService: sessionService,
		parserService:  parserService,
	}
}

func (s *ResourceService) GetResources(ctx context.Context, schoolUrl string, sessionID string) ([]*booking.Resource, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	if err := s.sessionService.SetSessionLanguage(ctx, schoolUrl, sessionID); err != nil {
		return nil, fmt.Errorf("failed to set session language: %w", err)
	}

	resourcesHTML, err := s.getResourcesHTML(ctx, schoolUrl, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources HTML: %w", err)
	}

	resources, err := s.parserService.ParseResources(resourcesHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resources: %w", err)
	}

	return resources, nil
}

func (s *ResourceService) GetBookedResources(ctx context.Context, schoolUrl string, sessionID string) ([]*booking.Booking, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	if err := s.sessionService.SetSessionLanguage(ctx, schoolUrl, sessionID); err != nil {
		return nil, fmt.Errorf("failed to set session language: %w", err)
	}

	resourcesHTML, err := s.getResourcesHTML(ctx, schoolUrl, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources HTML: %w", err)
	}

	resources, err := s.parserService.ParseResources(resourcesHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse resources: %w", err)
	}

	var allBookings []*booking.Booking

	for _, resource := range resources {
		bookingsHTML, err := s.getActiveResourceBookings(ctx, schoolUrl, sessionID, resource.ID)
		if err != nil {
			fmt.Printf("Failed to get bookings HTML for resource %s: %v\n", resource.ID, err)
			continue
		}

		bookings, err := s.parserService.ParsePersonalBookings(bookingsHTML, resource.ID)
		if err != nil {
			fmt.Printf("Failed to parse bookings for resource %s: %v\n", resource.ID, err)
			continue
		}
		allBookings = append(allBookings, bookings...)
	}

	return allBookings, nil
}

func (s *ResourceService) GetActiveResourceBookings(ctx context.Context, schoolUrl, sessionID, resourceID string) ([]*booking.Booking, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	if err := s.sessionService.SetSessionLanguage(ctx, schoolUrl, sessionID); err != nil {
		return nil, fmt.Errorf("failed to set session language: %w", err)
	}

	bookingsHTML, err := s.getActiveResourceBookings(ctx, schoolUrl, sessionID, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get personal bookings HTML: %w", err)
	}

	bookings, err := s.parserService.ParsePersonalBookings(bookingsHTML, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse personal bookings: %w", err)
	}

	return bookings, nil
}

func (s *ResourceService) GetAvailableResources(ctx context.Context, school, sessionID string, date time.Time, resourceID string) ([]*booking.AvailabilitySlot, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	html, err := s.getAvailableResourcesHTML(ctx, school, sessionID, date, resourceID)
	if err != nil {
		return nil, err
	}

	return s.parserService.ParseResourceAvailability(html, date)
}

func (s *ResourceService) BookResource(ctx context.Context, school, sessionID string, req *booking.BookingRequest, resourceId string) error {
	if err := s.sessionService.SetSessionLanguage(ctx, school, sessionID); err != nil {
		return fmt.Errorf("failed to set session language: %w", err)
	}

	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/ajax/ajax_resursbokning.jsp", strings.TrimSuffix(school, "/"))
	params := map[string]string{
		"op":        "boka",
		"datum":     req.Date.Format("06-01-02"),
		"flik":      resourceId,
		"id":        *req.Slot.LocationId,
		"typ":       *req.Slot.ResourceType,
		"intervall": *req.Slot.TimeSlotId,
		"moment":    "Booked via Tumble",
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to book resource: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	content := string(body)
	if content != "OK" {
		return s.handleBookingError(content, school, req, resourceId)
	}

	return nil
}

func (s *ResourceService) UnbookResource(ctx context.Context, school, sessionID, bookingID string) error {
	if err := s.sessionService.SetSessionLanguage(ctx, school, sessionID); err != nil {
		return fmt.Errorf("failed to set session language: %w", err)
	}

	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/ajax/ajax_resursbokning.jsp", strings.TrimSuffix(school, "/"))
	params := map[string]string{
		"op":         "avboka",
		"bokningsId": bookingID,
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to unbook resource: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	content := string(body)
	if content != "OK" {
		return s.handleUnbookingError(content, school, bookingID)
	}

	return nil
}

func (s *ResourceService) ConfirmBooking(ctx context.Context, school, sessionID, bookingID, resourceID string) error {
	if err := s.sessionService.SetSessionLanguage(ctx, school, sessionID); err != nil {
		return fmt.Errorf("failed to set session language: %w", err)
	}

	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/ajax/ajax_resursbokning.jsp", strings.TrimSuffix(school, "/"))
	params := map[string]string{
		"op":         "konfirmera",
		"flik":       resourceID,
		"bokningsId": bookingID,
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return fmt.Errorf("failed to confirm booking: %w", err)
	}
	defer response.Body.Close()

	return nil
}

func (s *ResourceService) getResourcesHTML(ctx context.Context, schoolUrl string, sessionID string) (string, error) {
	if err := s.sessionService.SetSessionLanguage(ctx, schoolUrl, sessionID); err != nil {
		return "", fmt.Errorf("failed to set session language: %w", err)
	}

	endpoint := fmt.Sprintf("%s/resursbokning.jsp", strings.TrimSuffix(schoolUrl, "/"))

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, map[string]string{})
	if err != nil {
		return "", fmt.Errorf("failed to get resources: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	htmlContent := string(body)
	if len(htmlContent) == 0 {
		return "", fmt.Errorf("received empty response from resources endpoint")
	}

	if strings.Contains(strings.ToLower(htmlContent), "användarnamn:") &&
		strings.Contains(strings.ToLower(htmlContent), "lösenord:") {
		return "", fmt.Errorf("session expired - redirected to login page")
	}

	return htmlContent, nil
}

func (s *ResourceService) getActiveResourceBookings(ctx context.Context, schoolUrl string, sessionID, resourceID string) (string, error) {
	if err := s.sessionService.SetSessionLanguage(ctx, schoolUrl, sessionID); err != nil {
		return "", fmt.Errorf("failed to set session language: %w", err)
	}

	date := time.Now()
	endpoint := fmt.Sprintf("%s/minaresursbokningar.jsp", strings.TrimSuffix(schoolUrl, "/"))
	params := map[string]string{
		"datum": date.Format("06-01-02"),
		"flik":  resourceID,
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return "", fmt.Errorf("failed to get active bookings: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

func (s *ResourceService) getAvailableResourcesHTML(ctx context.Context, school, sessionID string, date time.Time, resourceID string) (string, error) {
	if err := s.sessionService.SetSessionLanguage(ctx, school, sessionID); err != nil {
		return "", fmt.Errorf("failed to set session language: %w", err)
	}

	endpoint := fmt.Sprintf("%s/ajax/ajax_resursbokning.jsp", strings.TrimSuffix(school, "/"))
	params := map[string]string{
		"op":    "hamtaBokningar",
		"datum": date.Format("06-01-02"),
		"flik":  resourceID,
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *ResourceService) handleBookingError(content, school string, req *booking.BookingRequest, resourceId string) error {
	switch {
	case strings.Contains(content, "do not have permissions to book resources"):
		return fmt.Errorf("kronox failed to authorize the user credentials")
	case strings.Contains(content, "colliding resources"):
		return fmt.Errorf("couldn't book resource")
	case strings.Contains(content, "max number of bookings"):
		return fmt.Errorf("you have already created max number of bookings")
	default:
		return fmt.Errorf("something went wrong while booking resource. Details:\nschoolUrl: %s\ndate: %s\nresourceId: %s\nlocationId: %s\nresourceType: %s\ntimeSlotId: %s\nError: %s",
			school,
			req.Date.Format("02-01-06"),
			resourceId,
			*req.Slot.LocationId,
			*req.Slot.ResourceType,
			*req.Slot.TimeSlotId,
			content)
	}
}

func (s *ResourceService) handleUnbookingError(content, school, bookingID string) error {
	switch content {
	case "Din användare har inte rättigheter att skapa resursbokningar.":
		return fmt.Errorf("kronox failed to authorize the user credentials")
	case "Du kan inte radera resursbokningar där du inte är bokare eller deltagare":
		return fmt.Errorf("couldn't unbook resource. Details:\nschoolUrl: %s\nbookingId: %s", school, bookingID)
	default:
		return fmt.Errorf("something went wrong while unbooking resource. Details:\nschoolUrl: %s\nbookingId: %s\nError: %s", school, bookingID, content)
	}
}
