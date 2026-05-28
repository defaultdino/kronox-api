package middleware

import (
	"net/http"
	"strings"

	"github.com/defaultdino/kronox-api/pkg/kronox"
	"github.com/gin-gonic/gin"
)

func SchoolValidationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		schoolCode := c.Query("school")
		if schoolCode == "" {
			c.Next()
			return
		}

		cfg := kronox.GetSchoolConfig()
		school, exists := cfg.Schools[schoolCode]
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":           "invalid or unauthorized school",
				"allowed_schools": GetAllowedSchools(),
				"format":          "Use school code only (e.g., 'hkr')",
			})
			c.Abort()
			return
		}

		c.Set("school_code", schoolCode)
		c.Set("school_name", school.Name)
		c.Set("school_domain", school.Domain)
		c.Set("school_urls", school.URLs)

		urlMap := make(map[string]string)
		for _, u := range school.URLs {
			if strings.Contains(u, "schema.") {
				urlMap["schema"] = u
			}
			if strings.Contains(u, "kronox.") {
				urlMap["kronox"] = u
			}
			if strings.Contains(u, "webbschema.") {
				urlMap["webbschema"] = u
			}
		}
		c.Set("school_url_map", urlMap)

		c.Next()
	})
}

func GetSchoolCode(c *gin.Context) string {
	if code, ok := c.Get("school_code"); ok {
		return code.(string)
	}
	return ""
}

func GetSchoolName(c *gin.Context) string {
	if name, ok := c.Get("school_name"); ok {
		return name.(string)
	}
	return ""
}

func GetSchoolDomain(c *gin.Context) string {
	if domain, ok := c.Get("school_domain"); ok {
		return domain.(string)
	}
	return ""
}

func GetSchoolURLs(c *gin.Context) []string {
	if urls, ok := c.Get("school_urls"); ok {
		return urls.([]string)
	}
	return []string{}
}

func GetURLByType(c *gin.Context, urlType string) string {
	if urlMap, ok := c.Get("school_url_map"); ok {
		if u, ok := urlMap.(map[string]string)[urlType]; ok {
			return u
		}
	}
	return ""
}

func HasURLType(c *gin.Context, urlType string) bool {
	return GetURLByType(c, urlType) != ""
}

func GetAllowedSchools() []string {
	cfg := kronox.GetSchoolConfig()
	codes := make([]string, 0, len(cfg.Schools))
	for code := range cfg.Schools {
		codes = append(codes, code)
	}
	return codes
}

func GetSchoolInfo(c *gin.Context) *kronox.School {
	code := GetSchoolCode(c)
	if code == "" {
		return nil
	}
	cfg := kronox.GetSchoolConfig()
	if school, ok := cfg.Schools[code]; ok {
		return &school
	}
	return nil
}
