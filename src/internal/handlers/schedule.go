package handlers

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
	"github.com/tumble-for-kronox/kronox-api/pkg/middleware"
	"github.com/tumble-for-kronox/kronox-api/pkg/models"
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

// GetScheduleEvents godoc
// @Summary      Get schedule events
// @Description  Retrieve schedule events for one or more schedule IDs with optional language and date filtering
// @Tags         schedules
// @Accept       json
// @Produce      json
// @Param        schedule_ids  query     string  true   "Comma-separated list of schedule IDs"  example("schedule1,schedule2,schedule3")
// @Param        start_date    query     string  false  "Start date for filtering events (YYYY-MM-DD format)"  example("2024-01-15") format(date)
// @Param        school    query     string  true  "School that request pertains to"  example("hkr")
// @Param        url_index query     int     true  "Index of the school URL to use"  example(0)
// @Success      200          {object}  ScheduleEventsResponse  "List of schedule events"
// @Failure      400          {object}  ErrorResponse           "Missing required parameters, invalid date format, or invalid url_index"
// @Failure      500          {object}  ErrorResponse           "Failed to fetch or parse schedule data"
// @Router       /schedules [get]
func (h *ScheduleHandler) GetScheduleEvents(c *gin.Context) {
	scheduleIDsParam := c.Query("schedule_ids")
	if scheduleIDsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schedule_ids required (comma-separated)"})
		c.Abort()
		return
	}

	scheduleIDs := strings.Split(scheduleIDsParam, ",")
	for i, id := range scheduleIDs {
		scheduleIDs[i] = url.QueryEscape(id)
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

	schoolURL, err := GetSchoolURLFromIndex(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	scheduleXML, err := h.scheduleService.GetScheduleEvents(ctx, schoolURL, scheduleIDs, startDate)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch schedule from specified URL"})
		return
	}

	schoolCode := middleware.GetSchoolCode(c)

	if schoolCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not retrieve schoolCode"})
		return
	}

	events, err := h.parserService.ParseScheduleXML(schoolCode, scheduleIDs, scheduleXML)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse schedule XML"})
		return
	}

	if events == nil {
		events = []*models.Event{}
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

// ScheduleEvent represents a single schedule event
// @Description Schedule event information
type ScheduleEvent struct {
	ID         string `json:"id" example:"evt_123"`
	Title      string `json:"title" example:"Math Lecture"`
	StartTime  string `json:"startTime" example:"2024-01-15T09:00:00Z"`
	EndTime    string `json:"endTime" example:"2024-01-15T10:30:00Z"`
	Location   string `json:"location" example:"Room A101"`
	Instructor string `json:"instructor,omitempty" example:"Dr. Smith"`
	CourseCode string `json:"courseCode,omitempty" example:"MATH101"`
}

// ScheduleEventsResponse represents the response for schedule events
// @Description Response containing list of schedule events parsed from XML
type ScheduleEventsResponse struct {
	Events []ScheduleEvent `json:"events"`
}
