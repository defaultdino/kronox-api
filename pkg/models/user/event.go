package user

import "time"

type Event struct {
	Title string    `json:"title"`
	Type  string    `json:"type"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
