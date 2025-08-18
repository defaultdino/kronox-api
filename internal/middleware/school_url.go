// internal/middleware/school_validation.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var AllowedSchools = map[string]bool{
	"hkr": true, // Högskolan Kristianstad
	"mau": true, // Malmö universitet
	"oru": true, // Örebro universitet
	"ltu": true, // Luleå tekniska universitet
	"hig": true, // Högskolan i Gävle
	"sh":  true, // Södertörns högskola
	"hv":  true, // Högskolan Väst
	"hb":  true, // Högskolan i Borås
	"mdh": true, // Mälardalens universitet
}

func SchoolValidationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		schoolCode := c.Query("school")

		if schoolCode == "" {
			c.Next()
			return
		}

		if !AllowedSchools[schoolCode] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":           "invalid or unauthorized school",
				"allowed_schools": GetAllowedSchools(),
				"format":          "Use school code only (e.g., 'hkr')",
			})
			c.Abort()
			return
		}

		domain := schoolCode + ".se"
		c.Set("school_code", schoolCode)
		c.Set("school_domain", domain)
		c.Set("schema_url", "https://schema."+domain)
		c.Set("kronox_url", "https://kronox."+domain)
		c.Set("webbschema_url", "https://webbschema."+domain)

		c.Next()
	})
}

func GetSchemaURL(c *gin.Context) string {
	if url, exists := c.Get("schema_url"); exists {
		return url.(string)
	}
	return ""
}

func GetWebbschemaURL(c *gin.Context) string {
	if url, exists := c.Get("webbschema_url"); exists {
		return url.(string)
	}
	return ""
}

func GetKronoxURL(c *gin.Context) string {
	if url, exists := c.Get("kronox_url"); exists {
		return url.(string)
	}
	return ""
}

func GetSchoolCode(c *gin.Context) string {
	if code, exists := c.Get("school_code"); exists {
		return code.(string)
	}
	return ""
}

func GetAllowedSchools() []string {
	var schools []string
	for school := range AllowedSchools {
		schools = append(schools, school)
	}
	return schools
}
