package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
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

func (h *ScheduleHandler) GetSchedule(c *gin.Context) {
	schemaURL := middleware.GetSchemaURL(c)
	webbschemaURL := middleware.GetWebbschemaURL(c)
	kronoxURL := middleware.GetKronoxURL(c)

	if schemaURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "school required"})
		return
	}

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

	// automatically attempt fetching from schema -> webbschema -> kronox

	scheduleXML, err := h.scheduleService.GetSchedules(c.Request.Context(), schemaURL, scheduleIDs, language, startDate)
	if err != nil {
		scheduleXML, err = h.scheduleService.GetSchedules(c.Request.Context(), webbschemaURL, scheduleIDs, language, startDate)
		if err != nil {
			scheduleXML, err = h.scheduleService.GetSchedules(c.Request.Context(), kronoxURL, scheduleIDs, language, startDate)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch schedule from both schema and webbschema URLs"})
				return
			}
		}
	}

	events, err := h.parserService.ParseScheduleXML(scheduleXML)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse schedule XML"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (h *ScheduleHandler) SearchProgrammes(c *gin.Context) {
	schemaURL := middleware.GetSchemaURL(c)
	webbschemaURL := middleware.GetWebbschemaURL(c)

	if schemaURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "school required"})
		return
	}

	searchQuery := c.Query("q")
	if searchQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query 'q' required"})
		return
	}

	programmesHTML, err := h.scheduleService.GetProgrammes(c.Request.Context(), schemaURL, searchQuery)
	if err != nil {
		programmesHTML, err = h.scheduleService.GetProgrammes(c.Request.Context(), webbschemaURL, searchQuery)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch programmes from both schema and webbschema URLs"})
			return
		}
	}

	programmes, err := h.parserService.ParseProgrammes(programmesHTML)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse programmes HTML"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"programmes": programmes})
}
