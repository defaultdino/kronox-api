package resource

import (
	"time"
)

type Resource struct {
	ID             string                               `json:"id"`
	Name           string                               `json:"name"`
	TimeSlots      []*TimeSlot                          `json:"time_slots,omitempty"`
	Date           *time.Time                           `json:"date,omitempty"`
	LocationIDs    []string                             `json:"location_ids,omitempty"`
	Availabilities map[string]map[int]*AvailabilitySlot `json:"availabilities,omitempty"`
}
