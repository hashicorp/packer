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

// CreateBackendSetDetails The configuration details for creating a backend set in a load balancer.
// For more information on backend set configuration, see
// Managing Backend Sets (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/tasks/managingbackendsets.htm).
type CreateBackendSetDetails struct {
	HealthChecker *HealthCheckerDetails `mandatory:"true" json:"healthChecker"`

	// A friendly name for the backend set. It must be unique and it cannot be changed.
	// Valid backend set names include only alphanumeric characters, dashes, and underscores. Backend set names cannot
	// contain spaces. Avoid entering confidential information.
	// Example: `My_backend_set`
	Name *string `mandatory:"true" json:"name"`

	// The load balancer policy for the backend set. To get a list of available policies, use the
	// ListPolicies operation.
	// Example: `LEAST_CONNECTIONS`
	Policy *string `mandatory:"true" json:"policy"`

	Backends []BackendDetails `mandatory:"false" json:"backends"`

	SessionPersistenceConfiguration *SessionPersistenceConfigurationDetails `mandatory:"false" json:"sessionPersistenceConfiguration"`

	SslConfiguration *SslConfigurationDetails `mandatory:"false" json:"sslConfiguration"`
}

func (m CreateBackendSetDetails) String() string {
	return common.PointerString(m)
}
