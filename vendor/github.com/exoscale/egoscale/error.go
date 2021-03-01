package egoscale

import "errors"

// ErrNotFound represents an error indicating a non-existent resource.
var ErrNotFound = errors.New("resource not found")

// ErrTooManyFound represents an error indicating multiple results found for a single resource.
var ErrTooManyFound = errors.New("multiple resources found")

// ErrInvalidRequest represents an error indicating that the caller's request is invalid.
var ErrInvalidRequest = errors.New("invalid request")

// ErrAPIError represents an error indicating an API-side issue.
var ErrAPIError = errors.New("API error")
