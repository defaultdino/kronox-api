package booking

import (
	"fmt"
	"time"
)

type TimeSlot struct {
	ID   *int      `json:"id,omitempty"`
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

func (ts *TimeSlot) Duration() time.Duration {
	return ts.To.Sub(ts.From)
}

func (ts *TimeSlot) String() string {
	return fmt.Sprintf("%s-%s",
		ts.From.Format("15:04"),
		ts.To.Format("15:04"))
}

func (ts *TimeSlot) Equal(other *TimeSlot) bool {
	if ts == nil || other == nil {
		return ts == other
	}

	if ts.ID != nil && other.ID != nil {
		return *ts.ID == *other.ID
	}

	if ts.ID == nil && other.ID == nil {
		return ts.From.Equal(other.From) && ts.To.Equal(other.To)
	}

	return false
}

func (ts *TimeSlot) Hash() uint64 {
	if ts.ID != nil {
		return uint64(*ts.ID)
	}
	return uint64(ts.From.Unix() + ts.To.Unix())
}
