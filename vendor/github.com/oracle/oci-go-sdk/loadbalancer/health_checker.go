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

// HealthChecker The health check policy configuration.
// For more information, see Editing Health Check Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/Tasks/editinghealthcheck.htm).
type HealthChecker struct {

	// The backend server port against which to run the health check. If the port is not specified, the load balancer uses the
	// port information from the `Backend` object.
	// Example: `8080`
	Port *int `mandatory:"true" json:"port"`

	// The protocol the health check must use; either HTTP or TCP.
	// Example: `HTTP`
	Protocol *string `mandatory:"true" json:"protocol"`

	// A regular expression for parsing the response body from the backend server.
	// Example: `^(500|40[1348])$`
	ResponseBodyRegex *string `mandatory:"true" json:"responseBodyRegex"`

	// The status code a healthy backend server should return. If you configure the health check policy to use the HTTP protocol,
	// you can use common HTTP status codes such as "200".
	// Example: `200`
	ReturnCode *int `mandatory:"true" json:"returnCode"`

	// The interval between health checks, in milliseconds. The default is 10000 (10 seconds).
	// Example: `30000`
	IntervalInMillis *int `mandatory:"false" json:"intervalInMillis"`

	// The number of retries to attempt before a backend server is considered "unhealthy". Defaults to 3.
	// Example: `3`
	Retries *int `mandatory:"false" json:"retries"`

	// The maximum time, in milliseconds, to wait for a reply to a health check. A health check is successful only if a reply
	// returns within this timeout period. Defaults to 3000 (3 seconds).
	// Example: `6000`
	TimeoutInMillis *int `mandatory:"false" json:"timeoutInMillis"`

	// The path against which to run the health check.
	// Example: `/healthcheck`
	UrlPath *string `mandatory:"false" json:"urlPath"`
}

func (m HealthChecker) String() string {
	return common.PointerString(m)
}
