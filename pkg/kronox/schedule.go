package kronox

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/defaultdino/kronox-api/pkg/models"
)

type kronoxScheduleXML struct {
	XMLName         xml.Name         `xml:"schema"`
	Posts           []eventPost      `xml:"schemaPost"`
	ExplanationRows []explanationRow `xml:"forklaringstexter>forklaringsrader"`
}

func ParseScheduleXML(schoolCode string, scheduleIDs []string, xmlContent string) ([]*models.Event, error) {
	var doc kronoxScheduleXML
	if err := xml.Unmarshal([]byte(xmlContent), &doc); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	teachers := parseTeachers(doc.ExplanationRows)
	locations := parseLocations(doc.ExplanationRows)
	courses := parseCourses(doc.ExplanationRows)

	var events []*models.Event
	for _, post := range doc.Posts {
		event, err := parseEvent(schoolCode, scheduleIDs, post, teachers, locations, courses)
		if err != nil {
			continue
		}
		events = append(events, event)
	}
	return events, nil
}

type eventPost struct {
	Moment       string `xml:"moment"`
	ActivityType struct {
		ID   string `xml:"id,attr"`
		Name string `xml:"namn,attr"`
	} `xml:"aktivitetsTyp"`
	BookingID   string `xml:"bokningsId"`
	BookedDates struct {
		StartDateTime string `xml:"startDatumTid_iCal,attr"`
		EndDateTime   string `xml:"slutDatumTid_iCal,attr"`
	} `xml:"bokadeDatum"`
	LastModified string `xml:"senastAndradDatum_iCal"`
	ResourceRow  struct {
		ResourceNodes []struct {
			ResourceTypeID       string `xml:"resursTypId,attr"`
			ResourceID           string `xml:"resursId"`
			ResourceIDURLEncoded string `xml:"resursIdURLEncoded"`
		} `xml:"resursNod"`
	} `xml:"resursTrad"`
}

type explanationRow struct {
	Type string `xml:"typ,attr"`
	Rows []struct {
		Columns []struct {
			Header string `xml:"rubrik,attr"`
			Value  string `xml:",chardata"`
		} `xml:"kolumn"`
	} `xml:"rad"`
}

func parseCourses(rows []explanationRow) map[string]string {
	courses := make(map[string]string)
	for _, row := range rows {
		if row.Type != "UTB_KURSINSTANS_GRUPPER" {
			continue
		}
		for _, rad := range row.Rows {
			var id, name string
			for _, kolumn := range rad.Columns {
				v := strings.TrimSpace(kolumn.Value)
				switch kolumn.Header {
				case "Id":
					id = v
				case "KursNamn_SV":
					name = v
				}
			}
			if id != "" && name != "" {
				courses[id] = name
			}
		}
	}
	return courses
}

func parseTeachers(rows []explanationRow) map[string]*models.Teacher {
	teachers := make(map[string]*models.Teacher)
	for _, row := range rows {
		if row.Type != "RESURSER_SIGNATURER" {
			continue
		}
		for _, rad := range row.Rows {
			var id, first, last string
			for _, kolumn := range rad.Columns {
				v := strings.TrimSpace(kolumn.Value)
				switch kolumn.Header {
				case "Id":
					id = v
				case "ForNamn":
					first = v
				case "EfterNamn":
					last = v
				}
			}
			if id != "" {
				teachers[id] = &models.Teacher{ID: id, FirstName: first, LastName: last}
			}
		}
	}
	return teachers
}

func parseLocations(rows []explanationRow) map[string]*models.Location {
	locations := make(map[string]*models.Location)
	for _, row := range rows {
		if row.Type != "RESURSER_LOKALER" {
			continue
		}
		for _, rad := range row.Rows {
			var id, name, building, floor, maxSeats string
			for _, kolumn := range rad.Columns {
				v := strings.TrimSpace(kolumn.Value)
				switch kolumn.Header {
				case "Id":
					id = v
				case "Lokalnamn":
					name = v
				case "Hus":
					building = v
				case "Vaning":
					floor = v
				case "Antalplatser":
					maxSeats = v
				}
			}
			if id != "" {
				locations[id] = &models.Location{
					ID:       id,
					Name:     name,
					Building: building,
					Floor:    floor,
					MaxSeats: maxSeats,
				}
			}
		}
	}
	return locations
}

var titleHTMLTags = regexp.MustCompile(`<.*?>`)

func parseEvent(schoolCode string, scheduleIDs []string, post eventPost, teachers map[string]*models.Teacher, locations map[string]*models.Location, courses map[string]string) (*models.Event, error) {
	timeStart, err := time.Parse("20060102T150405Z", post.BookedDates.StartDateTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}
	timeEnd, err := time.Parse("20060102T150405Z", post.BookedDates.EndDateTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %w", err)
	}
	lastModified, _ := time.Parse("20060102T150405Z", post.LastModified)

	title := strings.TrimSpace(titleHTMLTags.ReplaceAllString(post.Moment, ""))

	var courseID, courseName string
	var eventTeachers []*models.Teacher
	var eventLocations []*models.Location
	var resourceScheduleIDs []string

	for _, node := range post.ResourceRow.ResourceNodes {
		cleanID := strings.TrimSpace(node.ResourceID)
		cleanID = strings.TrimPrefix(cleanID, "<![CDATA[")
		cleanID = strings.TrimSuffix(cleanID, "]]>")
		cleanID = strings.TrimSpace(cleanID)

		switch node.ResourceTypeID {
		case "UTB_KURSINSTANS_GRUPPER":
			courseID = cleanID
			if name, exists := courses[cleanID]; exists {
				courseName = name
			}
		case "RESURSER_SIGNATURER":
			if t, exists := teachers[cleanID]; exists {
				eventTeachers = append(eventTeachers, t)
			}
		case "RESURSER_LOKALER":
			if l, exists := locations[cleanID]; exists {
				eventLocations = append(eventLocations, l)
			}
		case "UTB_PROGRAMINSTANS_KLASSER":
			if node.ResourceIDURLEncoded != "" {
				resourceScheduleIDs = append(resourceScheduleIDs, node.ResourceIDURLEncoded)
			} else {
				resourceScheduleIDs = append(resourceScheduleIDs, cleanID)
			}
		}
	}

	primaryScheduleID := matchScheduleID(scheduleIDs, resourceScheduleIDs)
	if primaryScheduleID == "" && len(scheduleIDs) > 0 {
		primaryScheduleID = scheduleIDs[0]
	}

	if courseID == "" {
		courseID = "N/A"
	}
	if courseName == "" {
		courseName = "N/A"
	}

	return &models.Event{
		ID:           post.BookingID,
		ScheduleID:   primaryScheduleID,
		Title:        title,
		CourseID:     courseID,
		CourseName:   courseName,
		Teachers:     eventTeachers,
		From:         timeStart,
		To:           timeEnd,
		Locations:    eventLocations,
		LastModified: lastModified,
		IsSpecial:    false,
		SchoolCode:   schoolCode,
		Color:        "#4A90E2",
	}, nil
}

func matchScheduleID(requested, fromXML []string) string {
	for _, xmlID := range fromXML {
		for _, reqID := range requested {
			if dot := strings.Index(reqID, "."); dot != -1 {
				if reqID[dot+1:] == xmlID {
					return reqID
				}
			} else if reqID == xmlID {
				return reqID
			}
		}
	}
	return ""
}

func removeDuplicateWords(input string) string {
	seen := make(map[string]bool)
	re := regexp.MustCompile(`\w+`)
	return re.ReplaceAllStringFunc(input, func(word string) string {
		upper := strings.ToUpper(word)
		if seen[upper] {
			return ""
		}
		seen[upper] = true
		return word
	})
}
