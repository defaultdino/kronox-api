package parsers

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tumble-for-kronox/kronox-api/pkg/models/user"
)

func (s *service) ParseUserEvents(html string) (*user.EventsResponse, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	response := &user.EventsResponse{
		Registered:   []*user.AvailableUserEvent{},
		Unregistered: []*user.AvailableUserEvent{},
		Upcoming:     []*user.UpcomingUserEvent{},
	}

	doc.Find("html body div:nth-child(2) div:nth-child(4) div div:first-child div").Children().Each(func(i int, s *goquery.Selection) {
		if event := parseAvailableEvent(s, true); event != nil {
			response.Registered = append(response.Registered, event)
		}
	})

	doc.Find("html body div:nth-child(2) div:nth-child(4) div div:nth-child(2) div").Children().Each(func(i int, s *goquery.Selection) {
		if event := parseAvailableEvent(s, false); event != nil {
			response.Unregistered = append(response.Unregistered, event)
		}
	})

	doc.Find("html body div:nth-child(2) div:nth-child(4) div div:nth-child(3) div").Children().Each(func(i int, s *goquery.Selection) {
		if event := parseUpcomingEvent(s); event != nil {
			response.Upcoming = append(response.Upcoming, event)
		}
	})

	return response, nil
}

func parseAvailableEvent(s *goquery.Selection, isRegistered bool) *user.AvailableUserEvent {
	var supportAvailable, mustChooseLocation bool
	var id, participatorID, supportID *string
	var title, eventType, anonymousCode string
	var lastSignupDate, startTime, endTime time.Time

	titleElement := s.Find("b").First()
	if titleElement.Length() == 0 {
		return nil
	}
	title = strings.TrimSpace(titleElement.Text())

	s.Find("a").Each(func(i int, button *goquery.Selection) {
		onclick, exists := button.Attr("onclick")
		if !exists {
			return
		}

		onclick = strings.ToLower(onclick)

		if strings.Contains(onclick, "stod") {
			supportAvailable = true
			re := regexp.MustCompile(`visastod\('(.*?)','(.*?)'\)`)
			matches := re.FindStringSubmatch(onclick)
			if len(matches) == 3 {
				participatorID = &matches[1]
				supportID = &matches[2]
			}
			return
		}

		if strings.Contains(onclick, "avanmal") {
			re := regexp.MustCompile(`avanmal\(.*?, '(.*?)'\)`)
			matches := re.FindStringSubmatch(onclick)
			if len(matches) == 2 {
				id = &matches[1]
			}
			return
		}

		if strings.Contains(onclick, "anmal") {
			re := regexp.MustCompile(`anmal\('(.*?)',\s*(.*?)\)`)
			matches := re.FindStringSubmatch(onclick)
			if len(matches) == 3 {
				id = &matches[1]
				if strings.TrimSpace(matches[2]) == "true" {
					mustChooseLocation = true
				}
			}
		}
	})

	var dataNodes []*goquery.Selection
	s.Find("div").EachWithBreak(func(i int, div *goquery.Selection) bool {
		if strings.Contains(strings.ToLower(div.Text()), "test date") {
			allDivs := s.Find("div")
			for j := 0; j < 6 && i+j < allDivs.Length(); j++ {
				dataNodes = append(dataNodes, allDivs.Eq(i+j))
			}
			return false
		}
		return true
	})

	if len(dataNodes) < 5 {
		return nil
	}

	rawDate := extractRegexGroup(dataNodes[0].Text(), `Test Date\s*:\s*(.*)`)
	rawStartTime := extractRegexGroup(dataNodes[1].Text(), `Start\s*:\s*(.*)`)
	rawEndTime := extractRegexGroup(dataNodes[2].Text(), `End\s*:\s*(.*)`)
	rawLastSignupDate := extractRegexGroup(dataNodes[3].Text(), `Registration closes\s*:\s*(.*)`)
	eventType = extractRegexGroup(dataNodes[4].Text(), `Test Type\s*:\s*(.*)`)

	if isRegistered && len(dataNodes) > 5 {
		anonymousCode = extractRegexGroup(dataNodes[5].Text(), `Anonymous Id\s*:\s*(.*)`)
	}

	var err error
	lastSignupDate, err = time.Parse("2006-01-02 15:04:05", rawLastSignupDate)
	if err != nil {
		return nil
	}

	startTime, err = time.Parse("2006-01-02 15:04", rawDate+" "+rawStartTime)
	if err != nil {
		return nil
	}

	endTime, err = time.Parse("2006-01-02 15:04", rawDate+" "+rawEndTime)
	if err != nil {
		return nil
	}

	return &user.AvailableUserEvent{
		ID:                       id,
		ParticipatorID:           participatorID,
		SupportID:                supportID,
		AnonymousCode:            anonymousCode,
		IsRegistered:             isRegistered,
		SupportAvailable:         supportAvailable,
		RequiresChoosingLocation: mustChooseLocation,
		LastSignupDate:           lastSignupDate,
		Title:                    title,
		Type:                     eventType,
		EventStart:               startTime,
		EventEnd:                 endTime,
	}
}

func parseUpcomingEvent(s *goquery.Selection) *user.UpcomingUserEvent {
	var title, eventType string
	var firstSignupDate, startTime, endTime time.Time

	titleElement := s.Find("b").First()
	if titleElement.Length() == 0 {
		return nil
	}
	title = strings.TrimSpace(titleElement.Text())

	dataNodes := s.Find("div").Slice(1, 6)
	if dataNodes.Length() < 5 {
		return nil
	}

	rawDate := extractRegexGroup(dataNodes.Eq(0).Text(), `Test Date\s*:\s*(.*)`)
	rawStartTime := extractRegexGroup(dataNodes.Eq(1).Text(), `Start\s*:\s*(.*)`)
	rawEndTime := extractRegexGroup(dataNodes.Eq(2).Text(), `End\s*:\s*(.*)`)
	rawFirstSignupDate := extractRegexGroup(dataNodes.Eq(3).Text(), `Registration opens\s*:\s*(.*)`)
	eventType = extractRegexGroup(dataNodes.Eq(4).Text(), `Test Type\s*:\s*(.*)`)

	var err error
	firstSignupDate, err = time.Parse("2006-01-02 15:04:05", rawFirstSignupDate)
	if err != nil {
		return nil
	}

	startTime, err = time.Parse("2006-01-02 15:04", rawDate+" "+rawStartTime)
	if err != nil {
		return nil
	}

	endTime, err = time.Parse("2006-01-02 15:04", rawDate+" "+rawEndTime)
	if err != nil {
		return nil
	}

	return &user.UpcomingUserEvent{
		FirstSignupDate: firstSignupDate,
		Title:           title,
		Type:            eventType,
		EventStart:      startTime,
		EventEnd:        endTime,
	}
}

func extractRegexGroup(text, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
