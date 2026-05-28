package kronox

import (
	"encoding/json"
	"fmt"
	"os"
)

type School struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Domain  string   `json:"domain"`
	URLs    []string `json:"urls"`
	LogoUrl string   `json:"logoUrl"`
}

type SchoolsConfig struct {
	Schools map[string]School `json:"schools"`
}

const (
	EnvSchoolsJSON = "KRONOX_SCHOOLS_JSON"
	EnvSchoolsFile = "KRONOX_SCHOOLS_FILE"

	defaultSchoolsPath = ".well-known/schools.json"
)

var schoolConfig SchoolsConfig

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

func SetSchools(cfg SchoolsConfig) {
	schoolConfig = cfg
}

func GetSchoolConfig() SchoolsConfig {
	return schoolConfig
}

func (c SchoolsConfig) GetMaxURLIndex(schoolCode string) int {
	school, ok := c.Schools[schoolCode]
	if !ok || len(school.URLs) == 0 {
		return 0
	}
	return len(school.URLs) - 1
}

func loadFromFile(path, source string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read schools config (%s=%q): %w", source, path, err)
	}
	return loadFromJSON(data, fmt.Sprintf("%s=%s", source, path))
}

func loadFromJSON(data []byte, source string) error {
	var cfg SchoolsConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("parse schools JSON from %s: %w", source, err)
	}
	if len(cfg.Schools) == 0 {
		return fmt.Errorf("schools config from %s is empty", source)
	}
	schoolConfig = cfg
	return nil
}
