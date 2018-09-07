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

// LoadBalancerPolicy A policy that determines how traffic is distributed among backend servers.
// For more information on load balancing policies, see
// How Load Balancing Policies Work (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/Reference/lbpolicies.htm).
type LoadBalancerPolicy struct {

	// The name of a load balancing policy.
	// Example: 'LEAST_CONNECTIONS'
	Name *string `mandatory:"true" json:"name"`
}

func (m LoadBalancerPolicy) String() string {
	return common.PointerString(m)
}
