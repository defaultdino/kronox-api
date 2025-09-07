package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/handlers"
	"github.com/tumble-for-kronox/kronox-api/pkg/middleware"
)

func SetupResourceRoutes(api *gin.RouterGroup, ResourceHandler *handlers.ResourceHandler) {
	protected := []gin.HandlerFunc{
		middleware.SessionMiddleware(),
		middleware.SchoolValidationMiddleware(),
	}

	resources := api.Group("/resources")
	resources.Use(protected...)
	{
		resources.GET("/all", ResourceHandler.GetAllResources)
		resources.GET("/:resourceId/availability", ResourceHandler.GetAvailableResources)
		resources.GET("/:resourceId/bookings", ResourceHandler.GetActiveBookingsForResource)

		resources.POST("/:resourceId/book", ResourceHandler.BookResource)
		resources.DELETE("/:resourceId/unbook", ResourceHandler.UnbookResource)
		resources.PUT("/booking/:bookingId/confirm", ResourceHandler.ConfirmResourceBookingWithBody)
	}
}
