package parsers

import (
	"time"

	"github.com/tumble-for-kronox/kronox-api/pkg/models"
	"github.com/tumble-for-kronox/kronox-api/pkg/models/booking"
)

type ParserService interface {
	// Booking parsers
	ParseResources(html string) ([]*booking.Resource, error)
	ParsePersonalBookings(html string, resourceID string) ([]*booking.Booking, error)
	ParseResourceAvailability(html string, resourceDate time.Time) ([]*booking.AvailabilitySlot, error) // Fixed method name

	// Schedule parsers
	ParseScheduleXML(xmlContent string) ([]*models.Event, error)
	ParseProgrammes(html string) ([]*models.Programme, error)

	ParseUserLogin(html string) (*LoginInfo, error)
	ParseUserEvents(html string) (*EventsResponse, error)
}
