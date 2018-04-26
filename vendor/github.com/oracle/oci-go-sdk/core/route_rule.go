// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"github.com/oracle/oci-go-sdk/common"
)

// RouteRule A mapping between a destination IP address range and a virtual device to route matching
// packets to (a target).
type RouteRule struct {

	// A destination IP address range in CIDR notation. Matching packets will
	// be routed to the indicated network entity (the target).
	// Example: `0.0.0.0/0`
	CidrBlock *string `mandatory:"true" json:"cidrBlock"`

	// The OCID for the route rule's target. For information about the type of
	// targets you can specify, see
	// Route Tables (https://docs.us-phoenix-1.oraclecloud.com/Content/Network/Tasks/managingroutetables.htm).
	NetworkEntityId *string `mandatory:"true" json:"networkEntityId"`
}

func (m RouteRule) String() string {
	return common.PointerString(m)
}
