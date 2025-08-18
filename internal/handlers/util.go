package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/internal/middleware"
)

func AttemptOverSchoolURLs[T any](c *gin.Context, callback func(url string) (T, error)) (T, error) {
	school := middleware.GetSchoolInfo(c)
	var zero T
	for _, url := range school.URLs {
		result, err := callback(url)
		if err == nil {
			return result, nil
		}
	}
	return zero, nil
}

func AttemptOverSchoolURLsBool(c *gin.Context, callback func(url string) error) error {
	school := middleware.GetSchoolInfo(c)
	for _, url := range school.URLs {
		err := callback(url)
		return err
	}
	return nil
}
