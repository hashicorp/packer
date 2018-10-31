package opc

import (
	"net/http"
	"net/url"
)

// Config details the parameters needed to authenticate with Oracle Clouds API
type Config struct {
	Username       *string
	Password       *string
	IdentityDomain *string
	APIEndpoint    *url.URL
	MaxRetries     *int
	LogLevel       LogLevelType
	Logger         Logger
	HTTPClient     *http.Client
	UserAgent      *string
}

// NewConfig returns a blank config to populate with the neccessary fields to authenitcate with Oracle's API
func NewConfig() *Config {
	return &Config{}
}
