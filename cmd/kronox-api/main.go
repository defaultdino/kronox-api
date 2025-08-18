package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/app"
	"github.com/tumble-for-kronox/kronox-api/internal/handlers"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
)

func main() {
	app, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.Close()

	parserService := parsers.NewParserService()
	authService := services.NewAuthService(app, parserService)
	sessionService := services.NewSessionService(app)
	scheduleService := services.NewScheduleService(app)
	bookingService := services.NewBookingService(app, sessionService, parserService)

	authHandler := handlers.NewAuthHandler(authService)
	scheduleHandler := handlers.NewScheduleHandler(scheduleService, parserService)
	bookingHandler := handlers.NewBookingHandler(bookingService)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.InjectDependencies(app))

	setupRoutes(r, authHandler, scheduleHandler, bookingHandler)

	log.Printf("Server starting on port %s", app.Config.Port)
	if err := r.Run(":" + app.Config.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler, scheduleHandler *handlers.ScheduleHandler, bookingHandler *handlers.BookingHandler) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/routes", func(c *gin.Context) {
		routes := []map[string]string{
			{"method": "GET", "path": "/health", "description": "Health check"},
			{"method": "GET", "path": "/routes", "description": "List all available routes"},
			{"method": "POST", "path": "/api/v1/auth/login", "description": "Login to get session ID"},
			{"method": "GET", "path": "/api/v1/auth/validate", "description": "Validate session"},
			{"method": "GET", "path": "/api/v1/schedules/", "description": "Get schedule events"},
			{"method": "GET", "path": "/api/v1/schedules/programmes", "description": "Search programmes"},
			{"method": "GET", "path": "/api/v1/bookings/", "description": "Get user bookings (requires session)"},
			{"method": "POST", "path": "/api/v1/bookings/", "description": "Book a resource (requires session)"},
			{"method": "GET", "path": "/api/v1/bookings/availability", "description": "Get resource availability (requires session)"},
		}
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

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		auth.Use(middleware.SchoolValidationMiddleware())
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/validate", authHandler.ValidateSession)
		}

		schedules := api.Group("/schedules")
		schedules.Use(middleware.SchoolValidationMiddleware())
		{
			schedules.GET("/", scheduleHandler.GetSchedule)
			schedules.GET("/programmes", scheduleHandler.SearchProgrammes)
		}

		bookings := api.Group("/bookings")
		bookings.Use(middleware.SessionMiddleware())
		bookings.Use(middleware.SchoolValidationMiddleware())
		{
			bookings.GET("/", bookingHandler.GetUserBookings)
			bookings.POST("/", bookingHandler.BookResource)
			bookings.GET("/availability", bookingHandler.GetResourceAvailability)
		}
	}
}
