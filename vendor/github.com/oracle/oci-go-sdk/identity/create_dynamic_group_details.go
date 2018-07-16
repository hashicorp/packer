// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Identity and Access Management Service API
//
// APIs for managing users, groups, compartments, and policies.
//

package identity

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateDynamicGroupDetails Properties for creating a dynamic group.
type CreateDynamicGroupDetails struct {

	// The OCID of the tenancy containing the group.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The name you assign to the group during creation. The name must be unique across all groups
	// in the tenancy and cannot be changed.
	Name *string `mandatory:"true" json:"name"`

	// The matching rule to dynamically match an instance certificate to this dynamic group.
	// For rule syntax, see Managing Dynamic Groups (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Tasks/managingdynamicgroups.htm).
	MatchingRule *string `mandatory:"true" json:"matchingRule"`

	// The description you assign to the group during creation. Does not have to be unique, and it's changeable.
	Description *string `mandatory:"true" json:"description"`
}

func (m CreateDynamicGroupDetails) String() string {
	return common.PointerString(m)
}
