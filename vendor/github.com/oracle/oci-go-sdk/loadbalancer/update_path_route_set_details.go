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

// UpdatePathRouteSetDetails An updated set of path route rules that overwrites the existing set of rules.
type UpdatePathRouteSetDetails struct {

	// The set of path route rules.
	PathRoutes []PathRoute `mandatory:"true" json:"pathRoutes"`
}

func (m UpdatePathRouteSetDetails) String() string {
	return common.PointerString(m)
}
