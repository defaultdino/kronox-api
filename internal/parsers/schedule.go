package parsers

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/tumble-for-kronox/kronox-api/pkg/models"
)

type KronoxScheduleXML struct {
	XMLName         xml.Name         `xml:"schema"`
	Posts           []eventPost      `xml:"schemaPost"`
	ExplanationRows []explanationRow `xml:"forklaringstexter>forklaringsrader"`
}

func (s *service) ParseScheduleXML(xmlContent string) ([]*models.Event, error) {
	var scheduleXML KronoxScheduleXML
	if err := xml.Unmarshal([]byte(xmlContent), &scheduleXML); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w", err)
	}

	explanationRows := scheduleXML.ExplanationRows

	teachers := parseTeachers(explanationRows)
	locations := parseLocations(explanationRows)
	courses := parseCourses(explanationRows)

	var events []*models.Event

	for _, post := range scheduleXML.Posts {
		event, err := parseEvent(post, teachers, locations, courses)
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

func parseCourses(explanationRows []explanationRow) map[string]string {
	courses := make(map[string]string)

	for _, row := range explanationRows {
		if row.Type != "UTB_KURSINSTANS_GRUPPER" {
			continue
		}

		for _, rad := range row.Rows {
			var courseID, courseName string

			for _, kolumn := range rad.Columns {
				cleanValue := strings.TrimSpace(kolumn.Value)
				cleanValue = strings.Trim(cleanValue, " \t\n\r")

				switch kolumn.Header {
				case "Id":
					courseID = cleanValue
				case "KursNamn_SV":
					courseName = cleanValue
				}
			}

			if courseID != "" && courseName != "" {
				courses[courseID] = courseName
			}
		}
	}

	return courses
}

func parseTeachers(explanationRows []explanationRow) map[string]*models.Teacher {
	teachers := make(map[string]*models.Teacher)

	for _, row := range explanationRows {
		if row.Type != "RESURSER_SIGNATURER" {
			continue
		}

		for _, rad := range row.Rows {
			var teacherID, firstName, lastName string

			for _, kolumn := range rad.Columns {
				cleanValue := strings.TrimSpace(kolumn.Value)
				cleanValue = strings.Trim(cleanValue, " \t\n\r")

				switch kolumn.Header {
				case "Id":
					teacherID = cleanValue
				case "ForNamn":
					firstName = cleanValue
				case "EfterNamn":
					lastName = cleanValue
				}
			}

			if teacherID != "" {
				teachers[teacherID] = &models.Teacher{
					ID:        teacherID,
					FirstName: firstName,
					LastName:  lastName,
				}
			}
		}
	}

	return teachers
}

func parseLocations(explanationRows []explanationRow) map[string]*models.Location {
	locations := make(map[string]*models.Location)

	for _, row := range explanationRows {
		if row.Type != "RESURSER_LOKALER" {
			continue
		}

		for _, rad := range row.Rows {
			var locationID, name, building, floor, maxSeats string

			for _, kolumn := range rad.Columns {
				cleanValue := strings.TrimSpace(kolumn.Value)
				cleanValue = strings.Trim(cleanValue, " \t\n\r")

				switch kolumn.Header {
				case "Id":
					locationID = cleanValue
				case "Lokalnamn":
					name = cleanValue
				case "Hus":
					building = cleanValue
				case "Vaning":
					floor = cleanValue
				case "Antalplatser":
					maxSeats = cleanValue
				}
			}

			if locationID != "" {
				locations[locationID] = &models.Location{
					ID:       locationID,
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

func parseEvent(post eventPost, teachers map[string]*models.Teacher, locations map[string]*models.Location, courses map[string]string) (*models.Event, error) {
	timeStart, err := time.Parse("20060102T150405Z", post.BookedDates.StartDateTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse start time: %w", err)
	}

	timeEnd, err := time.Parse("20060102T150405Z", post.BookedDates.EndDateTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse end time: %w", err)
	}

	lastModified, _ := time.Parse("20060102T150405Z", post.LastModified)

	re := regexp.MustCompile(`<.*?>`)
	title := strings.TrimSpace(re.ReplaceAllString(post.Moment, ""))

	var courseID, courseName string
	var eventTeachers []*models.Teacher
	var eventLocations []*models.Location
	var scheduleIDs []string

	for _, node := range post.ResourceRow.ResourceNodes {
		cleanResourceID := strings.TrimSpace(node.ResourceID)
		cleanResourceID = strings.TrimPrefix(cleanResourceID, "<![CDATA[")
		cleanResourceID = strings.TrimSuffix(cleanResourceID, "]]>")
		cleanResourceID = strings.TrimSpace(cleanResourceID)

		cleanResourceIDURLEncoded := strings.TrimSpace(node.ResourceIDURLEncoded)

		switch node.ResourceTypeID {
		case "UTB_KURSINSTANS_GRUPPER":
			courseID = cleanResourceID
			if name, exists := courses[cleanResourceID]; exists {
				courseName = name
			}
		case "RESURSER_SIGNATURER":
			if teacher, exists := teachers[cleanResourceID]; exists {
				eventTeachers = append(eventTeachers, teacher)
			}
		case "RESURSER_LOKALER":
			if location, exists := locations[cleanResourceID]; exists {
				eventLocations = append(eventLocations, location)
			}
		case "UTB_PROGRAMINSTANS_KLASSER":
			if cleanResourceIDURLEncoded != "" {
				scheduleIDs = append(scheduleIDs, cleanResourceIDURLEncoded)
			} else {
				scheduleIDs = append(scheduleIDs, cleanResourceID)
			}
		}
	}

	var primaryScheduleID string
	if len(scheduleIDs) > 0 {
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
		ScheduleId:   primaryScheduleID,
		Title:        title,
		CourseId:     courseID,
		CourseName:   courseName,
		Teachers:     eventTeachers,
		From:         timeStart,
		To:           timeEnd,
		Locations:    eventLocations,
		LastModified: lastModified,
		IsSpecial:    post.ActivityType.ID == "A",
	}, nil
}

func removeDuplicateWords(input string) string {
	words := make(map[string]bool)
	re := regexp.MustCompile(`\w+`)

	return re.ReplaceAllStringFunc(input, func(word string) string {
		upperWord := strings.ToUpper(word)
		if words[upperWord] {
			return ""
		}
		words[upperWord] = true
		return word
	})
}
