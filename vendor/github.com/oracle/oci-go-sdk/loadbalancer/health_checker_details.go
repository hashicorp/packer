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

// HealthCheckerDetails The health check policy's configuration details.
type HealthCheckerDetails struct {

	// The protocol the health check must use; either HTTP or TCP.
	// Example: `HTTP`
	Protocol *string `mandatory:"true" json:"protocol"`

	// The interval between health checks, in milliseconds.
	// Example: `30000`
	IntervalInMillis *int `mandatory:"false" json:"intervalInMillis"`

	// The backend server port against which to run the health check. If the port is not specified, the load balancer uses the
	// port information from the `Backend` object.
	// Example: `8080`
	Port *int `mandatory:"false" json:"port"`

	// A regular expression for parsing the response body from the backend server.
	// Example: `^(500|40[1348])$`
	ResponseBodyRegex *string `mandatory:"false" json:"responseBodyRegex"`

	// The number of retries to attempt before a backend server is considered "unhealthy".
	// Example: `3`
	Retries *int `mandatory:"false" json:"retries"`

	// The status code a healthy backend server should return.
	// Example: `200`
	ReturnCode *int `mandatory:"false" json:"returnCode"`

	// The maximum time, in milliseconds, to wait for a reply to a health check. A health check is successful only if a reply
	// returns within this timeout period.
	// Example: `6000`
	TimeoutInMillis *int `mandatory:"false" json:"timeoutInMillis"`

	// The path against which to run the health check.
	// Example: `/healthcheck`
	UrlPath *string `mandatory:"false" json:"urlPath"`
}

func (m HealthCheckerDetails) String() string {
	return common.PointerString(m)
}
