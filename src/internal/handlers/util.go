package handlers

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tumble-for-kronox/kronox-api/pkg/middleware"
)

func AttemptOverSchoolURLs[T any](c *gin.Context, callback func(url string) (T, error)) (T, error) {
	school := middleware.GetSchoolInfo(c)
	var zero T

	if school == nil {
		return zero, fmt.Errorf("school info not available in context")
	}

	var errors []string
	
	for _, url := range school.URLs {
		fmt.Fprintf(gin.DefaultWriter, "Attempting school URL: %s\n", url)
		result, err := callback(url)
		if err == nil {
			fmt.Fprintf(gin.DefaultWriter, "Success with URL: %s\n", url)
			return result, nil
		}
		
		fmt.Fprintf(gin.DefaultWriter, "Failed with URL %s: %v\n", url, err)
		errors = append(errors, fmt.Sprintf("%s: %v", url, err))
	}
	
	return zero, fmt.Errorf("all school URLs failed: %s", strings.Join(errors, "; "))
}

func AttemptOverSchoolURLsBool(c *gin.Context, callback func(url string) error) error {
	school := middleware.GetSchoolInfo(c)

	if school == nil {
		return fmt.Errorf("school info not available in context")
	}

	var errors []string

	for _, url := range school.URLs {
		fmt.Fprintf(gin.DefaultWriter, "Attempting school URL: %s\n", url)
		err := callback(url)
		if err == nil {
			fmt.Fprintf(gin.DefaultWriter, "Success with URL: %s\n", url)
			return nil
		}

		fmt.Fprintf(gin.DefaultWriter, "Failed with URL %s: %v\n", url, err)
		errors = append(errors, fmt.Sprintf("%s: %v", url, err))
	}

	return fmt.Errorf("all school URLs failed: %s", strings.Join(errors, "; "))
}
