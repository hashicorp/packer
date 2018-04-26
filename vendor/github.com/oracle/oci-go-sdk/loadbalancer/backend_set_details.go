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

// BackendSetDetails The configuration details for a load balancer backend set.
// For more information on backend set configuration, see
// Managing Backend Sets (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/tasks/managingbackendsets.htm).
type BackendSetDetails struct {
	HealthChecker *HealthCheckerDetails `mandatory:"true" json:"healthChecker"`

	// The load balancer policy for the backend set. To get a list of available policies, use the
	// ListPolicies operation.
	// Example: `LEAST_CONNECTIONS`
	Policy *string `mandatory:"true" json:"policy"`

	Backends []BackendDetails `mandatory:"false" json:"backends"`

	SessionPersistenceConfiguration *SessionPersistenceConfigurationDetails `mandatory:"false" json:"sessionPersistenceConfiguration"`

	SslConfiguration *SslConfigurationDetails `mandatory:"false" json:"sslConfiguration"`
}

func (m BackendSetDetails) String() string {
	return common.PointerString(m)
}
