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

// UpdateGroupDetails The representation of UpdateGroupDetails
type UpdateGroupDetails struct {

	// The description you assign to the group. Does not have to be unique, and it's changeable.
	Description *string `mandatory:"false" json:"description"`
}

func (m UpdateGroupDetails) String() string {
	return common.PointerString(m)
}
