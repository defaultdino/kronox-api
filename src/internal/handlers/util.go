package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/defaultdino/kronox-api/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func AttemptOverSchoolURLs[T any](c *gin.Context, callback func(ctx context.Context, url string) (T, error)) (T, error) {
	school := middleware.GetSchoolInfo(c)
	var zero T

	if school == nil {
		return zero, fmt.Errorf("school info not available in context")
	}

	var errors []string

	for _, url := range school.URLs {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		result, err := callback(ctx, url)
		cancel()

		if err == nil {
			return result, nil
		}

		errors = append(errors, fmt.Sprintf("%s: %v", url, err))
	}

	return zero, fmt.Errorf("all school URLs failed: %s", strings.Join(errors, "; "))
}

func GetSchoolURLByIndex(c *gin.Context, urlIndex int) (string, error) {
	school := middleware.GetSchoolInfo(c)

	if school == nil {
		return "", fmt.Errorf("school info not available in context")
	}

	if urlIndex < 0 || urlIndex >= len(school.URLs) {
		return "", fmt.Errorf("url_index %d is out of bounds, school has %d URLs", urlIndex, len(school.URLs))
	}

	return school.URLs[urlIndex], nil
}

func GetSchoolURLFromIndex(c *gin.Context) (string, error) {
	urlIndexStr := c.Query("url_index")
	if urlIndexStr == "" {
		return "", fmt.Errorf("url_index query parameter is required")
	}

	urlIndex, err := strconv.Atoi(urlIndexStr)
	if err != nil {
		return "", fmt.Errorf("url_index must be a valid integer")
	}

	return GetSchoolURLByIndex(c, urlIndex)
}

func ExecuteWithSchoolURL(c *gin.Context, callback func(url string) error) error {
	schoolURL, err := GetSchoolURLFromIndex(c)
	if err != nil {
		return err
	}

	return callback(schoolURL)
}

// ErrorResponse represents a generic API error
// @Description Error response structure
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request"`
}
