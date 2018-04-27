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

// User An individual employee or system that needs to manage or use your company's Oracle Cloud Infrastructure
// resources. Users might need to launch instances, manage remote disks, work with your cloud network, etc. Users
// have one or more IAM Service credentials (ApiKey,
// UIPassword, and SwiftPassword).
// For more information, see User Credentials (https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/usercredentials.htm)). End users of your
// application are not typically IAM Service users. For conceptual information about users and other IAM Service
// components, see Overview of the IAM Service (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/overview.htm).
// These users are created directly within the Oracle Cloud Infrastructure system, via the IAM service.
// They are different from *federated users*, who authenticate themselves to the Oracle Cloud Infrastructure
// Console via an identity provider. For more information, see
// Identity Providers and Federation (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/federation.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access,
// see Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type User struct {

	// The OCID of the user.
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the tenancy containing the user.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The name you assign to the user during creation. This is the user's login for the Console.
	// The name must be unique across all users in the tenancy and cannot be changed.
	Name *string `mandatory:"true" json:"name"`

	// The description you assign to the user. Does not have to be unique, and it's changeable.
	Description *string `mandatory:"true" json:"description"`

	// Date and time the user was created, in the format defined by RFC3339.
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The user's current state. After creating a user, make sure its `lifecycleState` changes from CREATING to
	// ACTIVE before using it.
	LifecycleState UserLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Returned only if the user's `lifecycleState` is INACTIVE. A 16-bit value showing the reason why the user
	// is inactive:
	// - bit 0: SUSPENDED (reserved for future use)
	// - bit 1: DISABLED (reserved for future use)
	// - bit 2: BLOCKED (the user has exceeded the maximum number of failed login attempts for the Console)
	InactiveStatus *int `mandatory:"false" json:"inactiveStatus"`
}

func (m User) String() string {
	return common.PointerString(m)
}

// UserLifecycleStateEnum Enum with underlying type: string
type UserLifecycleStateEnum string

// Set of constants representing the allowable values for UserLifecycleState
const (
	UserLifecycleStateCreating UserLifecycleStateEnum = "CREATING"
	UserLifecycleStateActive   UserLifecycleStateEnum = "ACTIVE"
	UserLifecycleStateInactive UserLifecycleStateEnum = "INACTIVE"
	UserLifecycleStateDeleting UserLifecycleStateEnum = "DELETING"
	UserLifecycleStateDeleted  UserLifecycleStateEnum = "DELETED"
)

var mappingUserLifecycleState = map[string]UserLifecycleStateEnum{
	"CREATING": UserLifecycleStateCreating,
	"ACTIVE":   UserLifecycleStateActive,
	"INACTIVE": UserLifecycleStateInactive,
	"DELETING": UserLifecycleStateDeleting,
	"DELETED":  UserLifecycleStateDeleted,
}

// GetUserLifecycleStateEnumValues Enumerates the set of values for UserLifecycleState
func GetUserLifecycleStateEnumValues() []UserLifecycleStateEnum {
	values := make([]UserLifecycleStateEnum, 0)
	for _, v := range mappingUserLifecycleState {
		values = append(values, v)
	}
	return values
}
