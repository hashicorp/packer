package http

import (
	"time"
)

type mimeType string

const (
	mimeFormURLEncoded mimeType = "application/x-www-form-urlencoded"
	mimeJSON           mimeType = "application/json"
)

// DefaultHeaders defined default http headers
var DefaultHeaders = map[string]string{
	"Content-Type": string(mimeFormURLEncoded),
	// "X-SDK-VERSION": VERSION,
}

// DefaultTimeout is the default timeout of each request
var DefaultTimeout = 30 * time.Second
