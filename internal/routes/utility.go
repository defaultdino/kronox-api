package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
)

func SetupUtilityRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/routes", func(c *gin.Context) {
		routes := getAPIDocumentation()
		c.JSON(200, gin.H{"routes": routes})
	})

	r.GET("/schools", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"allowed_schools": middleware.GetAllowedSchools(),
			"format":          "school=<school_code>",
			"examples": []string{
				"school=hkr",
				"school=mau", 
				"school=oru",
			},
		})
	})
}

func getAPIDocumentation() []map[string]string {
	return []map[string]string{
		// Utility
		{"method": "GET", "path": "/health", "description": "Health check"},
		{"method": "GET", "path": "/routes", "description": "List all available routes"},
		{"method": "GET", "path": "/schools", "description": "List supported schools"},
		
		// Auth
		{"method": "POST", "path": "/api/v1/auth/login", "description": "Login to get session ID"},
		{"method": "GET", "path": "/api/v1/auth/validate", "description": "Validate session"},
		
		// Schedules
		{"method": "GET", "path": "/api/v1/schedules", "description": "Get schedule events"},
		{"method": "GET", "path": "/api/v1/schedules/programmes", "description": "Search programmes"},
		
		// Resources & Bookings
		{"method": "GET", "path": "/api/v1/resources", "description": "Get all bookable resources"},
		{"method": "GET", "path": "/api/v1/resources/{resourceId}/availability", "description": "Get resource availability"},
		{"method": "GET", "path": "/api/v1/resources/{resourceId}/bookings", "description": "Get bookings for specific resource"},
		
		// User Bookings
		{"method": "GET", "path": "/api/v1/bookings", "description": "Get user's all bookings"},
		{"method": "POST", "path": "/api/v1/bookings", "description": "Create new booking"},
		{"method": "DELETE", "path": "/api/v1/bookings/{bookingId}", "description": "Cancel booking"},
		{"method": "POST", "path": "/api/v1/bookings/{bookingId}/confirm", "description": "Confirm pending booking"},
	}
}