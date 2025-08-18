package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/app"
	"github.com/tumble-for-kronox/kronox-api/internal/handlers"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	"github.com/tumble-for-kronox/kronox-api/internal/routes"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
)

func main() {
	app, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer app.Close()

	services := initializeServices(app)
	
	handlers := initializeHandlers(services)

	r := setupRouter(app, handlers)

	log.Printf("Server starting on port %s", app.Config.Port)
	if err := r.Run(":" + app.Config.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

type Services struct {
	Parser    parsers.ParserService
	Auth      *services.AuthService
	Session   *services.SessionService
	Schedule  *services.ScheduleService
	Booking   *services.ResourceService
}

type Handlers struct {
	Auth     *handlers.AuthHandler
	Schedule *handlers.ScheduleHandler
	Booking  *handlers.ResourceHandler
}

func initializeServices(app *app.App) *Services {
	parserService := parsers.NewParserService()
	
	return &Services{
		Parser:   parserService,
		Auth:     services.NewAuthService(app, parserService),
		Session:  services.NewSessionService(app),
		Schedule: services.NewScheduleService(app),
		Booking:  services.NewResourceService(app, services.NewSessionService(app), parserService),
	}
}

func initializeHandlers(services *Services) *Handlers {
	return &Handlers{
		Auth:     handlers.NewAuthHandler(services.Auth),
		Schedule: handlers.NewScheduleHandler(services.Schedule, services.Parser),
		Booking:  handlers.NewResourceHandler(services.Booking),
	}
}

func setupRouter(app *app.App, handlers *Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.InjectDependencies(app))

	routes.SetupUtilityRoutes(r)
	
	api := r.Group("/api/v1")
	routes.SetupAuthRoutes(api, handlers.Auth)
	routes.SetupScheduleRoutes(api, handlers.Schedule)
	routes.SetupResourceRoutes(api, handlers.Booking)

	return r
}