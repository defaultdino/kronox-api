package user

import (
	"time"
)

type AvailableUserEvent struct {
	Title      string    `json:"title"`
	Type       string    `json:"type"`
	EventStart time.Time `json:"event_start"`
	EventEnd   time.Time `json:"event_end"`

	ID                       *string   `json:"id,omitempty"`
	ParticipatorID           *string   `json:"participator_id,omitempty"`
	SupportID                *string   `json:"support_id,omitempty"`
	AnonymousCode            string    `json:"anonymous_code"`
	IsRegistered             bool      `json:"is_registered"`
	SupportAvailable         bool      `json:"support_available"`
	RequiresChoosingLocation bool      `json:"requires_choosing_location"`
	LastSignupDate           time.Time `json:"last_signup_date"`
}
