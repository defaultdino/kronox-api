package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/handlers"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
)

func SetupProgrammeRoutes(api *gin.RouterGroup, programmeHandler *handlers.ProgrammeHandler) {
	programmes := api.Group("/programmes")
	programmes.Use(middleware.SchoolValidationMiddleware())
	{
		programmes.GET("", programmeHandler.SearchProgrammes)
	}
}
