package parsers

import (
	"time"

	"github.com/tumble-for-kronox/kronox-api/pkg/models"
	booking "github.com/tumble-for-kronox/kronox-api/pkg/models/resource"
	"github.com/tumble-for-kronox/kronox-api/pkg/models/user"
)

type ParserService interface {
	// Resource parsers
	ParseResources(html string) ([]*booking.Resource, error)
	ParsePersonalBookings(html string, resourceID string) ([]*booking.Booking, error)
	ParseResourceAvailability(html string, resourceDate time.Time) ([]*booking.AvailabilitySlot, error) // Fixed method name

	// Event parsers
	ParseUserEvents(html string) (*user.EventsResponse, error)

	// Schedule parsers
	ParseScheduleXML(schoolCode string, scheduleIDs []string, xmlContent string) ([]*models.Event, error)

	// Programme parsers
	ParseProgrammes(html string) ([]*models.Programme, error)

	// Login parsers
	ParseUserLogin(html string) (*LoginInfo, error)
}
