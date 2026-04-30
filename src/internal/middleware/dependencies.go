package middleware

import (
	"github.com/defaultdino/kronox-api/internal/app"
	"github.com/gin-gonic/gin"
)

func InjectDependencies(app *app.App) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Set("app", app)
		c.Next()
	})
}

func GetApp(c *gin.Context) *app.App {
	return c.MustGet("app").(*app.App)
}
