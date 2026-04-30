package routes

import (
	"github.com/defaultdino/kronox-api/internal/handlers"
	"github.com/defaultdino/kronox-api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func SetupScheduleRoutes(api *gin.RouterGroup, scheduleHandler *handlers.ScheduleHandler) {
	schedules := api.Group("/schedule")
	schedules.Use(middleware.SchoolValidationMiddleware())
	{
		schedules.GET("/events", scheduleHandler.GetScheduleEvents)
	}
}
