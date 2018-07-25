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

// UpdatePolicyDetails The representation of UpdatePolicyDetails
type UpdatePolicyDetails struct {

	// The description you assign to the policy. Does not have to be unique, and it's changeable.
	Description *string `mandatory:"false" json:"description"`

	// An array of policy statements written in the policy language. See
	// How Policies Work (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policies.htm) and
	// Common Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/commonpolicies.htm).
	Statements []string `mandatory:"false" json:"statements"`

	// The version of the policy. If null or set to an empty string, when a request comes in for authorization, the
	// policy will be evaluated according to the current behavior of the services at that moment. If set to a particular
	// date (YYYY-MM-DD), the policy will be evaluated according to the behavior of the services on that date.
	VersionDate *common.SDKTime `mandatory:"false" json:"versionDate"`
}

func (m UpdatePolicyDetails) String() string {
	return common.PointerString(m)
}
