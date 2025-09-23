package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
)

type School struct {
	Name   string   `toml:"name"`
	Domain string   `toml:"domain"`
	URLs   []string `toml:"urls"`
}

type Config struct {
	Schools map[string]School `toml:"schools"`
}

var (
	schoolConfig Config
	loadOnce     sync.Once
)

func GetSchoolConfig() Config {
	loadOnce.Do(func() {
		if err := LoadSchoolConfig("schools.toml"); err != nil {
			panic(err)
		}
	})
	return schoolConfig
}

func LoadSchoolConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := toml.Unmarshal(data, &schoolConfig); err != nil {
		return fmt.Errorf("failed to parse TOML config: %w", err)
	}

	return nil
}

func SchoolValidationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		schoolCode := c.Query("school")

		if schoolCode == "" {
			c.Next()
			return
		}

		cfg := GetSchoolConfig()
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
		for _, url := range school.URLs {
			if strings.Contains(url, "schema.") {
				urlMap["schema"] = url
			}
			if strings.Contains(url, "kronox.") {
				urlMap["kronox"] = url
			}
			if strings.Contains(url, "webbschema.") {
				urlMap["webbschema"] = url
			}
		}
		c.Set("school_url_map", urlMap)

		c.Next()
	})
}

func GetSchoolCode(c *gin.Context) string {
	if code, exists := c.Get("school_code"); exists {
		return code.(string)
	}
	return ""
}

func GetSchoolName(c *gin.Context) string {
	if name, exists := c.Get("school_name"); exists {
		return name.(string)
	}
	return ""
}

func GetSchoolDomain(c *gin.Context) string {
	if domain, exists := c.Get("school_domain"); exists {
		return domain.(string)
	}
	return ""
}

func GetSchoolURLs(c *gin.Context) []string {
	if urls, exists := c.Get("school_urls"); exists {
		return urls.([]string)
	}
	return []string{}
}

func GetURLByType(c *gin.Context, urlType string) string {
	if urlMap, exists := c.Get("school_url_map"); exists {
		if url, ok := urlMap.(map[string]string)[urlType]; ok {
			return url
		}
	}
	return ""
}

func HasURLType(c *gin.Context, urlType string) bool {
	return GetURLByType(c, urlType) != ""
}

func GetAllowedSchools() []string {
	var schools []string
	for school := range schoolConfig.Schools {
		schools = append(schools, school)
	}
	return schools
}

func GetSchoolInfo(c *gin.Context) *School {
	schoolCode := GetSchoolCode(c)
	if schoolCode == "" {
		return nil
	}

	if school, exists := schoolConfig.Schools[schoolCode]; exists {
		return &school
	}
	return nil
}
