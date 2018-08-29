// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Load Balancing Service API
//
// API for the Load Balancing Service
//

package loadbalancer

import (
	"github.com/oracle/oci-go-sdk/common"
)

// UpdateHealthCheckerDetails The health checker's configuration details.
type UpdateHealthCheckerDetails struct {

	// The interval between health checks, in milliseconds.
	// Example: `10000`
	IntervalInMillis *int `mandatory:"true" json:"intervalInMillis"`

	// The backend server port against which to run the health check.
	// Example: `8080`
	Port *int `mandatory:"true" json:"port"`

	// The protocol the health check must use; either HTTP or TCP.
	// Example: `HTTP`
	Protocol *string `mandatory:"true" json:"protocol"`

	// A regular expression for parsing the response body from the backend server.
	// Example: `^((?!false).|\s)*$`
	ResponseBodyRegex *string `mandatory:"true" json:"responseBodyRegex"`

	// The number of retries to attempt before a backend server is considered "unhealthy".
	// Example: `3`
	Retries *int `mandatory:"true" json:"retries"`

	// The status code a healthy backend server should return.
	// Example: `200`
	ReturnCode *int `mandatory:"true" json:"returnCode"`

	// The maximum time, in milliseconds, to wait for a reply to a health check. A health check is successful only if a reply
	// returns within this timeout period.
	// Example: `3000`
	TimeoutInMillis *int `mandatory:"true" json:"timeoutInMillis"`

	// The path against which to run the health check.
	// Example: `/healthcheck`
	UrlPath *string `mandatory:"false" json:"urlPath"`
}

func (m UpdateHealthCheckerDetails) String() string {
	return common.PointerString(m)
}
