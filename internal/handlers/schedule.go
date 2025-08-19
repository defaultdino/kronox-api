package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
)

type ScheduleHandler struct {
	scheduleService *services.ScheduleService
	parserService   parsers.ParserService
}

func NewScheduleHandler(scheduleService *services.ScheduleService, parserService parsers.ParserService) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleService: scheduleService,
		parserService:   parserService,
	}
}

// GetSchedule godoc
// @Summary      Get schedule events
// @Description  Retrieve schedule events for one or more schedule IDs with optional language and date filtering
// @Tags         schedules
// @Accept       json
// @Produce      json
// @Param        schedule_ids  query     string  true   "Comma-separated list of schedule IDs"  example("schedule1,schedule2,schedule3")
// @Param        language      query     string  false  "Language preference for the schedule"  example("en")
// @Param        start_date    query     string  false  "Start date for filtering events (YYYY-MM-DD format)"  example("2024-01-15") format(date)
// @Success      200          {object}  ScheduleEventsResponse  "List of schedule events"
// @Failure      400          {object}  ErrorResponse           "Missing required parameters or invalid date format"
// @Failure      500          {object}  ErrorResponse           "Failed to fetch or parse schedule data"
// @Router       /schedules [get]
func (h *ScheduleHandler) GetSchedule(c *gin.Context) {
	scheduleIDsParam := c.Query("schedule_ids")
	if scheduleIDsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schedule_ids required (comma-separated)"})
		return
	}
	scheduleIDs := strings.Split(scheduleIDsParam, ",")

	var language *string
	if lang := c.Query("language"); lang != "" {
		language = &lang
	}

	var startDate *time.Time
	if dateStr := c.Query("start_date"); dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			startDate = &parsed
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD"})
			return
		}
	}

	scheduleXML, err := AttemptOverSchoolURLs(c, func(url string) (string, error) {
		return h.scheduleService.GetSchedules(c.Request.Context(), url, scheduleIDs, language, startDate)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch schedule from both schema and webbschema URLs"})
		return
	}

	events, err := h.parserService.ParseScheduleXML(scheduleXML)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse schedule XML"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

// ScheduleEventsResponse represents the response for schedule events
// @Description Response containing list of schedule events parsed from XML
type ScheduleEventsResponse struct {
	Events interface{} `json:"events" example:"[{\"id\":\"evt_123\",\"title\":\"Math Lecture\",\"startTime\":\"2024-01-15T09:00:00Z\",\"endTime\":\"2024-01-15T10:30:00Z\",\"location\":\"Room A101\",\"instructor\":\"Dr. Smith\"}]"`
}
