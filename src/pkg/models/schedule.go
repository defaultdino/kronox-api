package models

import (
	"time"
)

type Event struct {
	ID           string      `json:"id" bson:"id"`
	ScheduleID   string      `json:"schedule_id" bson:"schedule_id"`
	Title        string      `json:"title" bson:"title"`
	CourseID     string      `json:"course_id" bson:"course_id"`
	CourseName   string      `json:"course_name" bson:"course_name"`
	Teachers     []*Teacher  `json:"teachers" bson:"teachers"`
	From         time.Time   `json:"from" bson:"from"`
	To           time.Time   `json:"to" bson:"to"`
	Locations    []*Location `json:"locations" bson:"locations"`
	LastModified time.Time   `json:"last_modified" bson:"last_modified"`
	IsSpecial    bool        `json:"is_special" bson:"is_special"`
}

type Teacher struct {
	ID        string `json:"id" bson:"id"`
	FirstName string `json:"first_name" bson:"firstname"`
	LastName  string `json:"last_name" bson:"last_name"`
}

type Location struct {
	ID       string `json:"id" bson:"id"`
	Name     string `json:"name" bson:"name"`
	Building string `json:"building" bson:"building"`
	Floor    string `json:"floor" bson:"floor"`
	MaxSeats string `json:"max_seats" bson:"max_seats"`
}
