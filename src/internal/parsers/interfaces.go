package parsers

import (
	"github.com/tumble-for-kronox/kronox-api/pkg/models"
)

type ParserService interface {
	ParseScheduleXML(schoolCode string, scheduleIDs []string, xmlContent string) ([]*models.Event, error)
	ParseProgrammes(html string) ([]*models.Programme, error)
}
