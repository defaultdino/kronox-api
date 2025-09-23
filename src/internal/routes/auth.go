package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/handlers"
	"github.com/tumble-for-kronox/kronox-api/pkg/middleware"
)

func SetupAuthRoutes(api *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	auth := api.Group("/auth")
	auth.Use(middleware.SchoolValidationMiddleware())
	{
		auth.POST("/login", authHandler.Login)
		auth.GET("/validate", authHandler.ValidateSession)
		auth.GET("/poll", authHandler.PollSession)
	}
}
