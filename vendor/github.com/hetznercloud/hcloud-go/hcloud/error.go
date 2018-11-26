package hcloud

import "fmt"

// ErrorCode represents an error code returned from the API.
type ErrorCode string

// Error codes returned from the API.
const (
	ErrorCodeServiceError      ErrorCode = "service_error"       // Generic server error
	ErrorCodeRateLimitExceeded ErrorCode = "rate_limit_exceeded" // Rate limit exceeded
	ErrorCodeUnknownError      ErrorCode = "unknown_error"       // Unknown error
	ErrorCodeNotFound          ErrorCode = "not_found"           // Resource not found
	ErrorCodeInvalidInput      ErrorCode = "invalid_input"       // Validation error

	// Deprecated error codes

	// The actual value of this error code is limit_reached. The new error code
	// rate_limit_exceeded for ratelimiting was introduced before Hetzner Cloud
	// launched into the public. To make clients using the old error code still
	// work as expected, we set the value of the old error code to that of the
	// new error code.
	ErrorCodeLimitReached = ErrorCodeRateLimitExceeded
)

// Error is an error returned from the API.
type Error struct {
	Code    ErrorCode
	Message string
	Details interface{}
}

func (e Error) Error() string {
	return fmt.Sprintf("%s (%s)", e.Message, e.Code)
}

// ErrorDetailsInvalidInput contains the details of an 'invalid_input' error.
type ErrorDetailsInvalidInput struct {
	Fields []ErrorDetailsInvalidInputField
}

// ErrorDetailsInvalidInputField contains the validation errors reported on a field.
type ErrorDetailsInvalidInputField struct {
	Name     string
	Messages []string
}

// IsError returns whether err is an API error with the given error code.
func IsError(err error, code ErrorCode) bool {
	apiErr, ok := err.(Error)
	return ok && apiErr.Code == code
}
