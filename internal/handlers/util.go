package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
)

func AttemptOverSchoolURLs[T any](c *gin.Context, callback func(url string) (T, error)) (T, error) {
	school := middleware.GetSchoolInfo(c)
	var zero T

	if school == nil {
		return zero, fmt.Errorf("school info not available in context")
	}

	for _, url := range school.URLs {
		result, err := callback(url)
		if err == nil {
			return result, nil
		}
	}
	return zero, fmt.Errorf("all school URLs failed")
}

func AttemptOverSchoolURLsBool(c *gin.Context, callback func(url string) error) error {
	school := middleware.GetSchoolInfo(c)

	if school == nil {
		return fmt.Errorf("school info not available in context")
	}

	for _, url := range school.URLs {
		err := callback(url)
		if err == nil {
			return nil
		}
	}
	return fmt.Errorf("all school URLs failed")
}
