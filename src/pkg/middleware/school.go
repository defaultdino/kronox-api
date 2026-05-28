package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type School struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Domain  string   `json:"domain"`
	URLs    []string `json:"urls"`
	LogoUrl string   `json:"logoUrl"`
}

type Config struct {
	Schools map[string]School `json:"schools"`
}

const (
	EnvSchoolsJSON = "KRONOX_SCHOOLS_JSON"
	EnvSchoolsFile = "KRONOX_SCHOOLS_FILE"

	defaultSchoolsPath = ".well-known/schools.json" // assuming the user is hosting their own API serving schools (like kron-api does)
)

var schoolConfig Config

func LoadSchools() error {
	if raw := os.Getenv(EnvSchoolsJSON); raw != "" {
		return loadFromJSON([]byte(raw), EnvSchoolsJSON)
	}
	path := os.Getenv(EnvSchoolsFile)
	source := EnvSchoolsFile
	if path == "" {
		path = defaultSchoolsPath
		source = "default path"
	}
	return loadFromFile(path, source)
}

func SetSchools(cfg Config) {
	schoolConfig = cfg
}

func loadFromFile(path, source string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read schools config (%s=%q): %w", source, path, err)
	}
	return loadFromJSON(data, fmt.Sprintf("%s=%s", source, path))
}

func loadFromJSON(data []byte, source string) error {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse schools JSON from %s: %w", source, err)
	}
	if len(cfg.Schools) == 0 {
		return fmt.Errorf("schools config from %s is empty", source)
	}
	schoolConfig = cfg
	return nil
}

func GetSchoolConfig() Config {
	return schoolConfig
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
