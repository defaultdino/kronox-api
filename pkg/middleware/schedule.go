package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func ExtractScheduleIDsMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		scheduleIDsParam := c.Query("schedule_ids")
		if scheduleIDsParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "schedule_ids required (comma-separated)"})
			c.Abort()
			return
		}

		rawScheduleIDs := strings.Split(scheduleIDsParam, ",")

		scheduleIDs := make([]string, len(rawScheduleIDs))
		for i, id := range rawScheduleIDs {
			scheduleIDs[i] = url.QueryEscape(id)
		}
		c.Set("schedule_ids", scheduleIDs)
		c.Next()
	})
}
