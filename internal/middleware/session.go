package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func SessionMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		var sessionID string

		if auth := c.GetHeader("Authorization"); auth != "" {
			if strings.HasPrefix(auth, "Bearer ") {
				sessionID = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if sessionID == "" {
			sessionID = c.GetHeader("X-Session-ID")
		}

		if sessionID == "" {
			if cookie, err := c.Request.Cookie("JSESSIONID"); err == nil {
				sessionID = cookie.Value
			}
		}

		if sessionID == "" {
			sessionID = c.Query("jsessionid")
		}

		if sessionID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing session ID (JSESSIONID)",
			})
			c.Abort()
			return
		}

		c.Set("session_id", sessionID)
		c.Next()
	})
}

func GetSessionID(c *gin.Context) (string, bool) {
	sessionID, exists := c.Get("session_id")
	if !exists {
		return "", false
	}

	if sid, ok := sessionID.(string); ok && sid != "" {
		return sid, true
	}

	return "", false
}
