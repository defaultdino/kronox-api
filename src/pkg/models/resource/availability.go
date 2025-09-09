package resource

type Availability string

const (
	Unavailable Availability = "UNAVAILABLE"
	Available   Availability = "AVAILABLE"
	Booked      Availability = "BOOKED"
)

func (a Availability) IsValid() bool {
	switch a {
	case Unavailable, Available, Booked:
		return true
	default:
		return false
	}
}

type AvailabilitySlot struct {
	Availability Availability `json:"availability"`
	LocationId   *string      `json:"location_id,omitempty"`
	ResourceType *string      `json:"resource_type,omitempty"`
	TimeSlotId   *string      `json:"time_slot_id,omitempty"`
}
