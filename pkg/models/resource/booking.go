package booking

import (
	"time"
)

type Booking struct {
	ResourceID         string     `json:"resource_id"`
	TimeSlot           *TimeSlot  `json:"time_slot"`
	LocationID         string     `json:"location_id"`
	ShowConfirmButton  bool       `json:"show_confirm_button"`
	ShowUnbookButton   bool       `json:"show_unbook_button"`
	ConfirmationOpen   *time.Time `json:"confirmation_open,omitempty"`
	ConfirmationClosed *time.Time `json:"confirmation_closed,omitempty"`
}

type BookingRequest struct {
	Date time.Time         `json:"date"`
	Slot *AvailabilitySlot `json:"slot"`
}

type ConfirmBookingRequest struct {
	ResourceID string `json:"resourceId"`
}
