package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetSchedule(scheduleId string, language string, startDate time.Time, c *gin.Context) {
	c.IndentedJSON(http.StatusOK, nil)
}
