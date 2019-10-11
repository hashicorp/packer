package http

import (
	"fmt"
)

// StatusError is the error for http status code >= 400
type StatusError struct {
	StatusCode int
	Message    string
}

func (e StatusError) Error() string {
	return fmt.Sprintf("http status %v error", e.StatusCode)
}

// NewStatusError will create a new status error
func NewStatusError(code int, message string) StatusError {
	return StatusError{
		StatusCode: code,
		Message:    message,
	}
}
