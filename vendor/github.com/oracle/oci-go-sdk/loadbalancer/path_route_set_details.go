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

// PathRouteSetDetails A set of path route rules.
type PathRouteSetDetails struct {

	// The set of path route rules.
	PathRoutes []PathRoute `mandatory:"true" json:"pathRoutes"`
}

func (m PathRouteSetDetails) String() string {
	return common.PointerString(m)
}
