package user

import (
	"time"
)

type UpcomingUserEvent struct {
	Title      string    `json:"title"`
	Type       string    `json:"type"`
	EventStart time.Time `json:"event_start"`
	EventEnd   time.Time `json:"event_end"`

	FirstSignupDate time.Time `json:"first_signup_date"`
}
