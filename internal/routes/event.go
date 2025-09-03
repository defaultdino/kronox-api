package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/handlers"
	"github.com/tumble-for-kronox/kronox-api/pkg/middleware"
)

func SetupEventRoutes(api *gin.RouterGroup, EventHandler *handlers.EventHandler) {
	protected := []gin.HandlerFunc{
		middleware.SessionMiddleware(),
		middleware.SchoolValidationMiddleware(),
	}

	events := api.Group("/events")
	support := events.Group("/support")
	events.Use(protected...)
	support.Use(protected...)
	{
		events.GET("/all", EventHandler.GetUserEvents)

		events.POST("/:eventId/register", EventHandler.RegisterUserEvent)
		events.DELETE("/:eventId/unregister", EventHandler.UnregisterUserEvent)

		support.POST("/:participatorId/:supportId", EventHandler.AddEventSupport)
		support.DELETE("/:eventId/:participatorId/:supportId", EventHandler.RemoveEventSupport)

		support.POST("/add/:participatorId", EventHandler.AddEventSupport)
		support.POST("/remove/:participatorId", EventHandler.RemoveEventSupport)
	}
}
