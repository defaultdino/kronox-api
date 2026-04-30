package routes

import (
	"github.com/defaultdino/kronox-api/internal/handlers"
	"github.com/defaultdino/kronox-api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func SetupProgrammeRoutes(api *gin.RouterGroup, programmeHandler *handlers.ProgrammeHandler) {
	programmes := api.Group("/programme/search")
	programmes.Use(middleware.SchoolValidationMiddleware())
	{
		programmes.GET("", programmeHandler.SearchProgrammes)
	}
}
