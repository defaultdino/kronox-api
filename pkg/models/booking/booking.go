package booking

import (
	"time"
)

type Booking struct {
	ID                 string     `json:"id"`
	ResourceID         string     `json:"resource_id"`
	TimeSlot           *TimeSlot  `json:"time_slot"`
	LocationID         string     `json:"location_id"`
	ShowConfirmButton  bool       `json:"show_confirm_button"`
	ShowUnbookButton   bool       `json:"show_unbook_button"`
	ConfirmationOpen   *time.Time `json:"confirmation_open,omitempty"`
	ConfirmationClosed *time.Time `json:"confirmation_closed,omitempty"`
}

type BookingRequest struct {
	ResourceId string           `json:"resource_id"`
	Date       time.Time        `json:"date"`
	Slot       AvailabilitySlot `json:"slot"`
}
