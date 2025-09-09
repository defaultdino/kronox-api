package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/handlers"
	"github.com/tumble-for-kronox/kronox-api/pkg/middleware"
)

func SetupResourceRoutes(api *gin.RouterGroup, resourceHandler *handlers.ResourceHandler) {
	protected := []gin.HandlerFunc{
		middleware.SessionMiddleware(),
		middleware.SchoolValidationMiddleware(),
	}

	resources := api.Group("/resources")

	resources.Use(protected...)
	{
		resources.GET("/all", resourceHandler.GetAllResources)
		resources.GET("/:resourceId/availability", resourceHandler.GetResourceAvailability)
		resources.GET("/:resourceId/bookings", resourceHandler.GetActiveBookingsForResource)
		resources.POST("/:resourceId/bookings", resourceHandler.BookResource)

		resources.GET("/bookings", resourceHandler.GetBookings)
		resources.DELETE("/bookings/:bookingId", resourceHandler.UnbookResource)
		resources.PUT("/bookings/:bookingId/confirm", resourceHandler.ConfirmResourceBooking)
	}
}
