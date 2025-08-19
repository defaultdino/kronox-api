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

	if err := r.Run(":" + app.Config.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

type Services struct {
	Parser   parsers.ParserService
	Auth     *services.AuthService
	Session  *services.SessionService
	Schedule *services.ScheduleService
	Resource *services.ResourceService
	Event    *services.EventService
}

type Handlers struct {
	Auth     *handlers.AuthHandler
	Schedule *handlers.ScheduleHandler
	Resource *handlers.ResourceHandler
	Event    *handlers.EventHandler
}

func initializeServices(app *app.App) *Services {
	parserService := parsers.NewParserService()
	sessionService := services.NewSessionService(app)

	return &Services{
		Parser:   parserService,
		Auth:     services.NewAuthService(app, parserService),
		Session:  services.NewSessionService(app),
		Schedule: services.NewScheduleService(app),
		Resource: services.NewResourceService(app, sessionService, parserService),
		Event:    services.NewEventService(app, sessionService, parserService),
	}
}

func initializeHandlers(services *Services) *Handlers {
	return &Handlers{
		Auth:     handlers.NewAuthHandler(services.Auth),
		Schedule: handlers.NewScheduleHandler(services.Schedule, services.Parser),
		Resource: handlers.NewResourceHandler(services.Resource),
		Event:    handlers.NewEventHandler(services.Event, services.Parser),
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
	routes.SetupResourceRoutes(api, handlers.Resource)
	routes.SetupEventRoutes(api, handlers.Event)

	return r
}
