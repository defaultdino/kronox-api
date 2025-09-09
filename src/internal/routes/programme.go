package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/handlers"
	"github.com/tumble-for-kronox/kronox-api/pkg/middleware"
)

func SetupProgrammeRoutes(api *gin.RouterGroup, programmeHandler *handlers.ProgrammeHandler) {
	programmes := api.Group("/programme/search")
	programmes.Use(middleware.SchoolValidationMiddleware())
	{
		programmes.GET("", programmeHandler.SearchProgrammes)
	}
}
