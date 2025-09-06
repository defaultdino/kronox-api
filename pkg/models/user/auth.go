package user

import booking "github.com/tumble-for-kronox/kronox-api/pkg/models/resource"

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// `omitempty` on events & bookings, since if these are not
// retrieved then they should not be in the returnd JSON whatsoever
// or stored in Mongo
type User struct {
	Name      string                `json:"name" bson:"name"`
	Username  string                `json:"username" bson:"username"`
	SessionID string                `json:"session_id" bson:"session_id"`
	Events    []*AvailableUserEvent `json:"events,omitempty" bson:"events,omitempty"`
	Bookings  []*booking.Booking    `json:"bookings,omitempty" bson:"bookings,omitempty"`
}
