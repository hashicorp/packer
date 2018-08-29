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

// UpdateDynamicGroupDetails Properties for updating a dynamic group.
type UpdateDynamicGroupDetails struct {

	// The description you assign to the dynamic group. Does not have to be unique, and it's changeable.
	Description *string `mandatory:"false" json:"description"`

	// The matching rule to dynamically match an instance certificate to this dynamic group.
	// For rule syntax, see Managing Dynamic Groups (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Tasks/managingdynamicgroups.htm).
	MatchingRule *string `mandatory:"false" json:"matchingRule"`
}

func (m UpdateDynamicGroupDetails) String() string {
	return common.PointerString(m)
}
