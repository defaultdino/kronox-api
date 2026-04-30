package main

import (
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/tumble-for-kronox/kronox-api/docs"
	"github.com/tumble-for-kronox/kronox-api/internal/app"
	"github.com/tumble-for-kronox/kronox-api/internal/handlers"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
	"github.com/tumble-for-kronox/kronox-api/internal/parsers"
	"github.com/tumble-for-kronox/kronox-api/internal/routes"
	"github.com/tumble-for-kronox/kronox-api/internal/services"
)

// @title           Kronox API
// @version         1.0
// @description     A RESTful Web API for managing Kronox web resources and schedules
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:5055
// @BasePath  /api/v1

func main() {
	app, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	services := initializeServices(app)
	handlers := initializeHandlers(services)

	r := setupRouter(app, handlers)

	if err := r.Run(":" + "5055"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

type Services struct {
	Parser    parsers.ParserService
	Schedule  *services.ScheduleService
	Programme *services.ProgrammeService
}

type Handlers struct {
	Schedule  *handlers.ScheduleHandler
	Programme *handlers.ProgrammeHandler
}

func initializeServices(app *app.App) *Services {
	parserService := parsers.NewParserService()

	return &Services{
		Parser:    parserService,
		Schedule:  services.NewScheduleService(app),
		Programme: services.NewProgrammeService(app),
	}
}

func initializeHandlers(services *Services) *Handlers {
	return &Handlers{
		Schedule:  handlers.NewScheduleHandler(services.Schedule, services.Parser),
		Programme: handlers.NewProgrammeHandler(services.Programme, services.Parser),
	}
}

func setupRouter(app *app.App, handlers *Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.InjectDependencies(app))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.SetupUtilityRoutes(r)

	api := r.Group("/api/v1")
	routes.SetupScheduleRoutes(api, handlers.Schedule)
	routes.SetupProgrammeRoutes(api, handlers.Programme)

	return r
}
