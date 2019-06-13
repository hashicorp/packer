/*
Package uerr is the error definition of service and sdk
*/
package uerr

// Error is the ucloud sdk error
type Error interface {
	error

	// name, should be client.xxx or server.xxx
	Name() string

	// retcode for server api error, retcode > 0 will cause an error
	Code() int

	// http status code, code >= 400 will case an error
	StatusCode() int

	// message for server api error
	Message() string

	// the origin error that sdk error caused from
	OriginError() error

	// if the error is retryable
	Retryable() bool
}

// NewRetryableError will wrap any error as a retryable error
func NewRetryableError(err error) Error {
	if e, ok := err.(ClientError); ok {
		e.retryable = true
		return e
	}

	if e, ok := err.(ServerError); ok {
		e.retryable = true
		return e
	}

	e := NewClientError(ErrUnexpected, err)
	e.retryable = true
	return e
}

// NewNonRetryableError will wrap any error as a non-retryable error
func NewNonRetryableError(err error) Error {
	if e, ok := err.(ClientError); ok {
		e.retryable = false
		return e
	}

	if e, ok := err.(ServerError); ok {
		e.retryable = false
		return e
	}

	e := NewClientError(ErrUnexpected, err)
	e.retryable = false
	return e
}
