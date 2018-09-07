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

// PathRouteSet A named set of path route rules. For more information, see
// Managing Request Routing (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/Tasks/managingrequest.htm).
type PathRouteSet struct {

	// The unique name for this set of path route rules. Avoid entering confidential information.
	// Example: `example_path_route_set`
	Name *string `mandatory:"true" json:"name"`

	// The set of path route rules.
	PathRoutes []PathRoute `mandatory:"true" json:"pathRoutes"`
}

func (m PathRouteSet) String() string {
	return common.PointerString(m)
}
