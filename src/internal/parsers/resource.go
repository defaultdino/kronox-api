package parsers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	booking "github.com/tumble-for-kronox/kronox-api/pkg/models/resource"

	"github.com/PuerkitoBio/goquery"
)

func (s *service) ParseResources(html string) ([]*booking.Resource, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
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
	_, exists := s.Attr("id")
	if !exists {
		return nil, fmt.Errorf("no booking ID found")
	}
	var showConfirmButton, showUnbookButton bool

	s.Find("a").Each(func(i int, link *goquery.Selection) {
		onclick, exists := link.Attr("onclick")
		text := strings.TrimSpace(link.Text())

		if exists && strings.Contains(onclick, "avboka(") {
			showUnbookButton = true
		}
		if exists && strings.Contains(onclick, "konfirmera(") {
			showConfirmButton = true
		}
		if text == "Confirm" || text == "Bekräfta" {
			showConfirmButton = true
		}
		if text == "Cancel booking" || text == "Avboka" {
			showUnbookButton = true
		}
	})

	firstDiv := s.Find("div").Eq(0)

	dateAnchor := firstDiv.Find("a").FilterFunction(func(i int, s *goquery.Selection) bool {
		onclick, _ := s.Attr("onclick")
		return strings.Contains(onclick, "satt_datum")
	}).First()

	dateText := strings.TrimSpace(dateAnchor.Text())

	var textParts []string
	firstDiv.Contents().Each(func(i int, node *goquery.Selection) {
		if goquery.NodeName(node) == "#text" {
			text := strings.TrimSpace(node.Text())
			if text != "" {
				textParts = append(textParts, text)
			}
		}
	})

	var timeText string
	timeRegex := regexp.MustCompile(`(\d{1,2}:\d{2}\s*-\s*\d{1,2}:\d{2})`)
	for _, part := range textParts {
		if matches := timeRegex.FindStringSubmatch(part); len(matches) > 1 {
			timeText = strings.TrimSpace(matches[1])
			break
		}
	}

	if timeText == "" {
		fullText := strings.TrimSpace(firstDiv.Text())
		fullText = strings.ReplaceAll(fullText, "Cancel booking", "")
		fullText = strings.ReplaceAll(fullText, "Avboka", "")
		fullText = strings.ReplaceAll(fullText, "Confirm", "")
		fullText = strings.ReplaceAll(fullText, "Bekräfta", "")

		matches := timeRegex.FindStringSubmatch(fullText)
		if len(matches) > 1 {
			timeText = strings.TrimSpace(matches[1])
		}
	}

	if timeText == "" {
		return nil, fmt.Errorf("could not extract time from booking node")
	}

	locationText := strings.TrimSpace(firstDiv.Find("b").Text())
	locationText = strings.TrimSpace(strings.TrimPrefix(locationText, "&nbsp;"))
	locationParts := strings.Split(locationText, ",")
	var locationID string
	if len(locationParts) >= 2 {
		locationID = strings.TrimSpace(locationParts[1])
	} else if len(locationParts) == 1 {
		locationID = strings.TrimSpace(locationParts[0])
	}

	timeParts := strings.Split(timeText, " - ")
	if len(timeParts) != 2 {
		return nil, fmt.Errorf("invalid time format: %s", timeText)
	}

	timeParts[0] = strings.TrimSpace(timeParts[0])
	timeParts[1] = strings.TrimSpace(timeParts[1])

	from, err := time.Parse("06-01-02 15:04", dateText+" "+timeParts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time '%s %s': %w", dateText, timeParts[0], err)
	}

	to, err := time.Parse("06-01-02 15:04", dateText+" "+timeParts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time '%s %s': %w", dateText, timeParts[1], err)
	}

	var confirmationOpen, confirmationClosed *time.Time
	secondDiv := s.Find("div").Eq(1)
	confirmationText := strings.TrimSpace(secondDiv.Find("span.error").Text())
	if confirmationText != "" {
		confirmationText = strings.Replace(confirmationText, "Måste bekräftas mellan ", "", 1)
		confirmationText = strings.Replace(confirmationText, "Must be confirmed between ", "", 1)

		confirmParts := strings.Split(confirmationText, " - ")
		if len(confirmParts) == 2 {
			if confirmFrom, err := time.Parse("06-01-02 15:04", dateText+" "+strings.TrimSpace(confirmParts[0])); err == nil {
				confirmationOpen = &confirmFrom
			}
			if confirmTo, err := time.Parse("06-01-02 15:04", dateText+" "+strings.TrimSpace(confirmParts[1])); err == nil {
				confirmationClosed = &confirmTo
			}
		}
	}

	timeSlot := &booking.TimeSlot{
		From: from,
		To:   to,
	}

	return &booking.Booking{
		ResourceID:         resourceID,
		TimeSlot:           timeSlot,
		LocationID:         locationID,
		ShowConfirmButton:  showConfirmButton,
		ShowUnbookButton:   showUnbookButton,
		ConfirmationOpen:   confirmationOpen,
		ConfirmationClosed: confirmationClosed,
	}, nil
}

type ResourceAvailabilityData struct {
	TimeSlots   []*booking.TimeSlot
	LocationIDs []string
	Slots       map[string]map[int]*booking.AvailabilitySlot
}

func (s *service) ParseResourceAvailability(html string, resourceDate time.Time) (*ResourceAvailabilityData, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	data := &ResourceAvailabilityData{
		Slots: make(map[string]map[int]*booking.AvailabilitySlot),
	}

	doc.Find("tr:first-child td b").Each(func(i int, s *goquery.Selection) {
		timeRange := strings.TrimSpace(s.Text())
		if timeRange == "" {
			return
		}

		parts := strings.Split(timeRange, " - ")
		if len(parts) != 2 {
			return
		}

		id := i
		timeSlot := &booking.TimeSlot{
			ID:   &id,
			From: parseTimeWithDate(parts[0], resourceDate),
			To:   parseTimeWithDate(parts[1], resourceDate),
		}
		data.TimeSlots = append(data.TimeSlots, timeSlot)
	})

	doc.Find("tr").Slice(1, goquery.ToEnd).Each(func(rowIndex int, row *goquery.Selection) {
		locationID := strings.TrimSpace(row.Find("td:first-child b").Text())
		if locationID == "" {
			return
		}

		data.LocationIDs = append(data.LocationIDs, locationID)
		data.Slots[locationID] = make(map[int]*booking.AvailabilitySlot)

		row.Find("td").Slice(1, goquery.ToEnd).Each(func(colIndex int, cell *goquery.Selection) {
			if colIndex >= len(data.TimeSlots) {
				return
			}

			class, _ := cell.Attr("class")
			classNames := strings.Split(class, " ")
			if len(classNames) == 0 {
				return
			}

			var availability booking.Availability
			var locationPtr, resourceType *string
			var timeSlotID *int

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
							locationPtr = &matches[1]
							resourceType = &matches[2]

							if timeSlotIdInt, err := strconv.Atoi(matches[3]); err == nil {
								timeSlotID = &timeSlotIdInt
							}
						}
					}
				}
			default:
				return
			}

			slot := &booking.AvailabilitySlot{
				Availability: availability,
				LocationId:   locationPtr,
				ResourceType: resourceType,
				TimeSlotId:   timeSlotID,
			}

			timeSlotId := *data.TimeSlots[colIndex].ID
			data.Slots[locationID][timeSlotId] = slot
		})
	})

	return data, nil
}

func parseTimeWithDate(timeStr string, baseDate time.Time) time.Time {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}
	}

	return time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(),
		t.Hour(), t.Minute(), 0, 0, baseDate.Location())
}
