package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/handlers"
	"github.com/tumble-for-kronox/kronox-api/pkg/middleware"
)

func SetupScheduleRoutes(api *gin.RouterGroup, scheduleHandler *handlers.ScheduleHandler) {
	schedules := api.Group("/schedule")
	schedules.Use(middleware.SchoolValidationMiddleware())
	{
		schedules.GET("/events", scheduleHandler.GetScheduleEvents)
	}
}
