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

// HostnameDetails The details of a hostname resource associated with a load balancer.
type HostnameDetails struct {

	// A virtual hostname. For more information about virtual hostname string construction, see
	// Managing Request Routing (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/Tasks/managingrequest.htm#routing).
	// Example: `app.example.com`
	Hostname *string `mandatory:"true" json:"hostname"`

	// The name of the hostname resource.
	// Example: `example_hostname_001`
	Name *string `mandatory:"true" json:"name"`
}

func (m HostnameDetails) String() string {
	return common.PointerString(m)
}
