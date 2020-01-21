package uerr

import (
	"fmt"
	"net"
	"strings"
)

var (
	// ErrInvalidRequest is the error for invalid request took from user
	ErrInvalidRequest = "client.InvalidRequestError"

	// ErrSendRequest is the error for sending request
	ErrSendRequest = "client.SendRequestError"

	// ErrNetwork is the error for any network error caused by client or server network environment
	// ErrNetwork can be caused by net errors of golang
	// ErrNetwork is retryable
	ErrNetwork = "client.NetworkError"

	// ErrUnexpected is the error for any unexpected error
	ErrUnexpected = "client.UnexpectedError"

	ErrCredentialExpired = "client.CredentialExpiredError"
)

// ClientError is the ucloud common errorfor server response
type ClientError struct {
	err       error
	name      string
	retryable bool
}

func (e ClientError) Error() string {
	return fmt.Sprintf("sdk:\n[%s] %s", e.name, e.err.Error())
}

// NewClientError will return a new instance of ClientError
func NewClientError(name string, err error) ClientError {
	return ClientError{
		name:      name,
		err:       err,
		retryable: isRetryableName(name),
	}
}

// Name will return error name
func (e ClientError) Name() string {
	return e.name
}

// Code will return server code
func (e ClientError) Code() int {
	return -1
}

// StatusCode will return http status code
func (e ClientError) StatusCode() int {
	return 0
}

// Message will return message
func (e ClientError) Message() string {
	return e.err.Error()
}

// OriginError will return the origin error that caused by
func (e ClientError) OriginError() error {
	return e.err
}

// Retryable will return if the error is retryable
func (e ClientError) Retryable() bool {
	return e.name == ErrNetwork || e.retryable
}

// IsNetworkError will check if the error raise from network problem
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}
	_, isNetError := err.(net.Error)
	if isNetError {
		return true
	}
	return strings.HasPrefix(err.Error(), "net/http: request canceled")
}

func isRetryableName(name string) bool {
	switch name {
	case ErrNetwork:
		return true
	default:
		return false
	}
}
