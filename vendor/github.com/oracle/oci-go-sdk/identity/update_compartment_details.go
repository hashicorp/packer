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

// UpdateCompartmentDetails The representation of UpdateCompartmentDetails
type UpdateCompartmentDetails struct {

	// The description you assign to the compartment. Does not have to be unique, and it's changeable.
	Description *string `mandatory:"false" json:"description"`

	// The new name you assign to the compartment. The name must be unique across all compartments in the tenancy.
	Name *string `mandatory:"false" json:"name"`
}

func (m UpdateCompartmentDetails) String() string {
	return common.PointerString(m)
}
