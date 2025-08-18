package parsers

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/tumble-for-kronox/kronox-api/pkg/models/booking"

	"github.com/PuerkitoBio/goquery"
)

func (s *service) ParseResources(html string) ([]*booking.Resource, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	if isSessionExpired(doc) {
		return nil, fmt.Errorf("session expired or invalid credentials")
	}

	var resources []*booking.Resource

	doc.Find("ul.menu").Eq(1).Find("li").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a")
		href, exists := link.Attr("href")
		if !exists || href == "" {
			return
		}

		re := regexp.MustCompile(`flik=(.*)`)
		matches := re.FindStringSubmatch(href)
		if len(matches) < 2 {
			return
		}

		resourceID := strings.TrimSpace(matches[1])
		resourceName := strings.TrimSpace(s.Find("a em b").Text())

		if resourceID != "" && resourceName != "" {
			resources = append(resources, &booking.Resource{
				ID:   resourceID,
				Name: resourceName,
			})
		}
	})

	return resources, nil
}

func (s *service) ParsePersonalBookings(html string, resourceID string) ([]*booking.Booking, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	if isSessionExpired(doc) {
		return nil, fmt.Errorf("session expired or invalid credentials")
	}

	var bookings []*booking.Booking

	doc.Find("#minabokningar div[id]").Each(func(i int, s *goquery.Selection) {
		booking, err := parseBookingNode(s, resourceID)
		if err == nil && booking != nil {
			bookings = append(bookings, booking)
		}
	})

	return bookings, nil
}

func parseBookingNode(s *goquery.Selection, resourceID string) (*booking.Booking, error) {
	id, exists := s.Attr("id")
	if !exists {
		return nil, fmt.Errorf("no booking ID found")
	}
	bookingID := strings.TrimPrefix(strings.TrimSpace(id), "post_")

	var showConfirmButton, showUnbookButton bool
	s.Find("div:first-child div:first-child a").Each(func(i int, link *goquery.Selection) {
		text := strings.TrimSpace(link.Text())
		switch text {
		case "Confirm":
			showConfirmButton = true
		case "Cancel booking":
			showUnbookButton = true
		}
	})

	dateText := strings.TrimSpace(s.Find("div:first-child a").Text())

	// Fix: Use FilterFunction instead of Filter
	timeText := strings.TrimSpace(s.Find("div:first-child").Contents().FilterFunction(func(i int, sel *goquery.Selection) bool {
		return goquery.NodeName(sel) == "#text"
	}).Text())

	// Extract location ID
	locationText := strings.TrimSpace(s.Find("div:first-child b").Text())
	locationParts := strings.Split(locationText, ",")
	var locationID string
	if len(locationParts) > 0 {
		locationID = strings.TrimSpace(locationParts[len(locationParts)-1])
	}

	timeParts := strings.Split(timeText, " - ")
	if len(timeParts) != 2 {
		return nil, fmt.Errorf("invalid time format: %s", timeText)
	}

	from, err := time.Parse("06-01-02 15:04", dateText+" "+timeParts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}

	to, err := time.Parse("06-01-02 15:04", dateText+" "+timeParts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %w", err)
	}

	var confirmationOpen, confirmationClosed *time.Time
	confirmationText := strings.TrimSpace(s.Find("div:nth-child(2) span").Text())
	if confirmationText != "" {
		confirmationText = strings.Replace(confirmationText, "Must be confirmed between ", "", 1)
		confirmParts := strings.Split(confirmationText, " - ")
		if len(confirmParts) == 2 {
			if confirmFrom, err := time.Parse("06-01-02 15:04", dateText+" "+confirmParts[0]); err == nil {
				confirmationOpen = &confirmFrom
			}
			if confirmTo, err := time.Parse("06-01-02 15:04", dateText+" "+confirmParts[1]); err == nil {
				confirmationClosed = &confirmTo
			}
		}
	}

	timeSlot := &booking.TimeSlot{
		From: from,
		To:   to,
	}

	return &booking.Booking{
		ID:                 bookingID,
		ResourceID:         resourceID,
		TimeSlot:           timeSlot,
		LocationID:         locationID,
		ShowConfirmButton:  showConfirmButton,
		ShowUnbookButton:   showUnbookButton,
		ConfirmationOpen:   confirmationOpen,
		ConfirmationClosed: confirmationClosed,
	}, nil
}

func (s *service) ParseResourceAvailability(html string, resourceDate time.Time) ([]*booking.AvailabilitySlot, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	if isSessionExpired(doc) {
		return nil, fmt.Errorf("session expired or invalid credentials")
	}

	var slots []*booking.AvailabilitySlot

	var timeSlots []*booking.TimeSlot
	doc.Find("tr:first-child td b").Each(func(i int, s *goquery.Selection) {
		timeRange := strings.TrimSpace(s.Text())
		if timeRange == "" {
			return
		}

		parts := strings.Split(timeRange, " - ")
		if len(parts) != 2 {
			return
		}

		timeSlot := &booking.TimeSlot{
			ID:   &i,
			From: parseTimeWithDate(parts[0], resourceDate),
			To:   parseTimeWithDate(parts[1], resourceDate),
		}
		timeSlots = append(timeSlots, timeSlot)
	})

	var locationIDs []string
	doc.Find("tr").Slice(1, goquery.ToEnd).Each(func(i int, s *goquery.Selection) {
		locationID := strings.TrimSpace(s.Find("td:first-child b").Text())
		if locationID != "" {
			locationIDs = append(locationIDs, locationID)
		}
	})

	doc.Find("tr").Slice(1, goquery.ToEnd).Each(func(rowIndex int, row *goquery.Selection) {
		if rowIndex >= len(locationIDs) {
			return
		}

		row.Find("td").Slice(1, goquery.ToEnd).Each(func(colIndex int, cell *goquery.Selection) {
			if colIndex >= len(timeSlots) {
				return
			}

			class, _ := cell.Attr("class")
			classNames := strings.Split(class, " ")
			if len(classNames) == 0 {
				return
			}

			var availability booking.Availability
			var locationID, resourceType, timeSlotID *string

			switch classNames[0] {
			case "grupprum-passerad":
				availability = booking.Unavailable
			case "grupprum-upptagen":
				availability = booking.Booked
			case "grupprum-ledig":
				availability = booking.Available

				if link := cell.Find("a"); link.Length() > 0 {
					onclick, exists := link.Attr("onclick")
					if exists {
						re := regexp.MustCompile(`boka\('(.*?)','(.*?)','(.*?)','(.*?)'\)`)
						matches := re.FindStringSubmatch(onclick)
						if len(matches) == 5 {
							locationID = &matches[1]
							resourceType = &matches[2]
							timeSlotID = &matches[3]
						}
					}
				}
			default:
				return
			}

			slot := &booking.AvailabilitySlot{
				Availability: availability,
				LocationId:   locationID,
				ResourceType: resourceType,
				TimeSlotId:   timeSlotID,
			}
			slots = append(slots, slot)
		})
	})

	return slots, nil
}

func parseTimeWithDate(timeStr string, baseDate time.Time) time.Time {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}
	}

	return time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		t.Hour(), t.Minute(), 0, 0, baseDate.Location())
}
