package main

import (
	"log"

	_ "github.com/defaultdino/kronox-api/docs"
	"github.com/defaultdino/kronox-api/internal/app"
	"github.com/defaultdino/kronox-api/internal/handlers"
	"github.com/defaultdino/kronox-api/internal/middleware"
	"github.com/defaultdino/kronox-api/internal/parsers"
	"github.com/defaultdino/kronox-api/internal/routes"
	"github.com/defaultdino/kronox-api/internal/services"
	sharedmw "github.com/defaultdino/kronox-api/pkg/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	if err := sharedmw.LoadSchools(); err != nil {
		log.Fatalf("Failed to load schools config: %v\n\n"+
			"Set %s (inline JSON), %s (path to a JSON file), or place schools.json at %s relative to cwd.",
			err, sharedmw.EnvSchoolsJSON, sharedmw.EnvSchoolsFile, ".well-known/schools.json")
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

	gin.SetMode(gin.ReleaseMode)

	r.Use(gin.Recovery())
	r.Use(middleware.InjectDependencies(app))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.SetupUtilityRoutes(r)

	api := r.Group("/api/v1")
	routes.SetupScheduleRoutes(api, handlers.Schedule)
	routes.SetupProgrammeRoutes(api, handlers.Programme)

	return r
}
