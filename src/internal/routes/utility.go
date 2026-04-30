package routes

import (
	"github.com/defaultdino/kronox-api/pkg/middleware"
	"github.com/gin-gonic/gin"
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
			"format":          "school=<school_code>&url_index=<index>",
			"examples": []string{
				"school=hkr&url_index=0",
				"school=mau&url_index=1",
				"school=oru&url_index=0",
			},
			"note": "url_index specifies which URL to use for the school (0-based index)",
		})
	})

	r.GET("/schools/:schoolCode/urls", func(c *gin.Context) {
		schoolCode := c.Param("schoolCode")

		// Create a temporary context with the school parameter to use the middleware functions
		tempContext := &gin.Context{}
		tempContext.Params = gin.Params{{Key: "school", Value: schoolCode}}
		tempContext.Request = c.Request
		tempContext.Set("school_code", schoolCode)

		// Load school config and get URLs
		cfg := middleware.GetSchoolConfig()
		if school, exists := cfg.Schools[schoolCode]; exists {
			urls := make([]map[string]interface{}, len(school.URLs))
			for i, url := range school.URLs {
				urls[i] = map[string]interface{}{
					"index": i,
					"url":   url,
				}
			}
			c.JSON(200, gin.H{
				"school": schoolCode,
				"urls":   urls,
				"usage":  "Use url_index parameter with the index value (0-based)",
			})
		} else {
			c.JSON(404, gin.H{
				"error":           "school not found",
				"allowed_schools": middleware.GetAllowedSchools(),
			})
		}
	})
}

func getAPIDocumentation() []map[string]string {
	return []map[string]string{
		// Utility
		{"method": "GET", "path": "/health", "description": "Health check"},
		{"method": "GET", "path": "/routes", "description": "List all available routes"},
		{"method": "GET", "path": "/schools", "description": "List supported schools"},

		// Schedules
		{"method": "GET", "path": "/api/v1/schedule/events", "description": "Get schedule events"},
		{"method": "GET", "path": "/api/v1/programme/search", "description": "Search programmes"},
	}
}
