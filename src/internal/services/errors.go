package services

import (
	"errors"
	"strings"
)

var (
	ErrSessionNotFound = errors.New("session not found or expired")
	ErrSessionExpired  = errors.New("session expired - redirected to login page")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrInvalidSession  = errors.New("invalid or expired session")
)

func IsAuthError(err error) bool {
	return errors.Is(err, ErrSessionNotFound) ||
		errors.Is(err, ErrSessionExpired) ||
		errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrInvalidSession) ||
		strings.Contains(err.Error(), "session not found") ||
		strings.Contains(err.Error(), "session expired") ||
		strings.Contains(err.Error(), "unauthorized") ||
		strings.Contains(err.Error(), "invalid session")
}
