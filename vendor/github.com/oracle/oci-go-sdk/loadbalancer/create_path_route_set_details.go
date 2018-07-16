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

// CreatePathRouteSetDetails A named set of path route rules to add to the load balancer.
type CreatePathRouteSetDetails struct {

	// The name for this set of path route rules. It must be unique and it cannot be changed. Avoid entering
	// confidential information.
	// Example: `example_path_route_set`
	Name *string `mandatory:"true" json:"name"`

	// The set of path route rules.
	PathRoutes []PathRoute `mandatory:"true" json:"pathRoutes"`
}

func (m CreatePathRouteSetDetails) String() string {
	return common.PointerString(m)
}
