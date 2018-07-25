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

// UpdateRouteTableDetails The representation of UpdateRouteTableDetails
type UpdateRouteTableDetails struct {

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The collection of rules used for routing destination IPs to network devices.
	RouteRules []RouteRule `mandatory:"false" json:"routeRules"`
}

func (m UpdateRouteTableDetails) String() string {
	return common.PointerString(m)
}
