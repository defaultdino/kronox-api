package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
)

type ProgrammeHandler struct {
	programmeService *services.ProgrammeService
	parserService    parsers.ParserService
}

func NewProgrammeHandler(programmeService *services.ProgrammeService, parserService parsers.ParserService) *ProgrammeHandler {
	return &ProgrammeHandler{
		programmeService: programmeService,
		parserService:    parserService,
	}
}

// SearchProgrammes godoc
// @Summary      Search for programmes
// @Description  Search for academic programmes across multiple school URLs using a search query
// @Tags         programmes
// @Accept       json
// @Produce      json
// @Param        q        query     string  true  "Search query for programmes"
// @Param        school   query     string  true  "School that request pertains to"  example("hkr")
// @Param        url_index query     int     true  "Index of the school URL to use"  example(0)
// @Success      200     {object}  ProgrammesListResponse  "List of programmes matching search criteria"
// @Failure      400     {object}  ErrorResponse           "Search query required or invalid url_index"
// @Failure      500     {object}  ErrorResponse           "Failed to fetch or parse programmes"
// @Router       /programmes [get]
func (h *ProgrammeHandler) SearchProgrammes(c *gin.Context) {
	searchQuery := c.Query("q")
	if searchQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query 'q' required"})
		return
	}

	schoolURL, err := GetSchoolURLFromIndex(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	programmesHTML, err := h.programmeService.GetProgrammes(ctx, schoolURL, searchQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch programmes from specified URL"})
		return
	}

	programmes, err := h.parserService.ParseProgrammes(programmesHTML)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse programmes HTML"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"programmes": programmes})
}

// Programme represents a single programme
// @Description Academic programme information
type Programme struct {
	ID          string `json:"id" example:"prog_123"`
	Name        string `json:"name" example:"Computer Science"`
	Code        string `json:"code" example:"CS101"`
	Credits     int    `json:"credits" example:"180"`
	Description string `json:"description,omitempty" example:"Bachelor's degree in Computer Science"`
}

// ProgrammesResponse represents the response for programme search
// @Description Response containing list of programmes matching the search criteria
type ProgrammesResponse struct {
	Programmes []Programme `json:"programmes"`
}
