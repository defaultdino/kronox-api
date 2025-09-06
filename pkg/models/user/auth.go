package user

import booking "github.com/tumble-for-kronox/kronox-api/pkg/models/resource"

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// stored in mongo
type User struct {
	Name     string                `json:"name" bson:"name"`
	Username string                `json:"username" bson:"username"`
	Events   []*AvailableUserEvent `json:"events" bson:"events"`
	Bookings []*booking.Booking    `json:"bookings" bson:"bookings"`
}
