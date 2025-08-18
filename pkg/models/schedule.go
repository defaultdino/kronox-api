package models

import (
	"time"
)

type Event struct {
	ID           string      `json:"id"`
	ScheduleId   string      `json:"schedule_id"`
	Title        string      `json:"title"`
	CourseId     string      `json:"course_id"`
	CourseName   string      `json:"course_name"`
	Teachers     []*Teacher  `json:"teachers"`
	From         time.Time   `json:"from"`
	To           time.Time   `json:"to"`
	Locations    []*Location `json:"locations"`
	LastModified time.Time   `json:"last_modified"`
	IsSpecial    bool        `json:"is_special"`
}

type Teacher struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Location struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Building string `json:"building"`
	Floor    string `json:"floor"`
	MaxSeats string `json:"max_seats"`
}
