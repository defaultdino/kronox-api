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
	"github.com/tumble-for-kronox/kronox-api/pkg/models/booking"
)

type BookingService struct {
	app            *app.App
	sessionService *SessionService
	parserService  parsers.ParserService
}

func NewBookingService(app *app.App, sessionService *SessionService, parserService parsers.ParserService) *BookingService {
	return &BookingService{
		app:            app,
		sessionService: sessionService,
		parserService:  parserService,
	}
}

func (s *BookingService) GetUserBookings(ctx context.Context, school, sessionID string) ([]*booking.Booking, error) {
	if err := s.sessionService.SetSessionLanguage(ctx, school, sessionID); err != nil {
		return nil, fmt.Errorf("failed to set session language: %w", err)
	}

	resources, err := s.getResources(ctx, school, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}

	var allBookings []*booking.Booking

	for _, resource := range resources {
		bookings, err := s.getPersonalBookingsForResource(ctx, school, sessionID, resource.ID)
		if err != nil {
			continue
		}
		allBookings = append(allBookings, bookings...)
	}

	return allBookings, nil
}

func (s *BookingService) GetResourceAvailability(ctx context.Context, school, sessionID string, date time.Time, resourceID string) ([]*booking.AvailabilitySlot, error) {
	if err := s.sessionService.SetSessionLanguage(ctx, school, sessionID); err != nil {
		return nil, fmt.Errorf("failed to set session language: %w", err)
	}

	html, err := s.getResourceAvailabilityHTML(ctx, school, sessionID, date, resourceID)
	if err != nil {
		return nil, err
	}

	return s.parserService.ParseResourceAvailability(html, date)
}

func (s *BookingService) BookResource(ctx context.Context, school, sessionID string, req *booking.BookingRequest) error {
	if err := s.sessionService.SetSessionLanguage(ctx, school, sessionID); err != nil {
		return fmt.Errorf("failed to set session language: %w", err)
	}

	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/ajax/ajax_resursbokning.jsp", strings.TrimSuffix(school, "/"))
	params := map[string]string{
		"op":        "boka",
		"datum":     req.Date.Format("06-01-02"),
		"flik":      req.ResourceId,
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
		return s.handleBookingError(content, req)
	}

	return nil
}

func (s *BookingService) getResources(ctx context.Context, school, sessionID string) ([]*booking.Resource, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/resursbokning.jsp", strings.TrimSuffix(school, "/"))
	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, map[string]string{})
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return s.parserService.ParseResources(string(body))
}

func (s *BookingService) getPersonalBookingsForResource(ctx context.Context, school, sessionID, resourceID string) ([]*booking.Booking, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

	endpoint := fmt.Sprintf("%s/minaresursbokningar.jsp", strings.TrimSuffix(school, "/"))
	params := map[string]string{
		"datum": time.Now().Format("06-01-02"),
		"flik":  resourceID,
	}

	response, err := s.app.KronoxClient.SendRequest(ctx, http.MethodGet, endpoint, params)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return s.parserService.ParsePersonalBookings(string(body), resourceID)
}

func (s *BookingService) handleBookingError(content string, req *booking.BookingRequest) error {
	switch {
	case strings.Contains(content, "do not have permissions"):
		return fmt.Errorf("user not authorized to book resources")
	case strings.Contains(content, "colliding resources"):
		return fmt.Errorf("booking collision occurred")
	case strings.Contains(content, "max number of bookings"):
		return fmt.Errorf("maximum bookings reached")
	default:
		return fmt.Errorf("booking failed: %s", content)
	}
}

func (s *BookingService) getResourceAvailabilityHTML(ctx context.Context, school, sessionID string, date time.Time, resourceID string) (string, error) {
	ctx = context.WithValue(ctx, sessionIDKey, sessionID)

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
