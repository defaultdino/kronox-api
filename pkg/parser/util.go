package parser

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// Standardized function to extract schedule_ids in a URL encoded manner,
// not stripping characters crucial to the identification of a schedule resource
func ExtractScheduleIDs(c *gin.Context, scheduleIDsParam string) []string {
	rawScheduleIDs := strings.Split(scheduleIDsParam, ",")

	scheduleIDs := make([]string, len(rawScheduleIDs))
	for i, id := range rawScheduleIDs {
		scheduleIDs[i] = url.QueryEscape(id)
	}
	return scheduleIDs
}
