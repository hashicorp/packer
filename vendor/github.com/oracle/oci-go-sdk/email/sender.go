// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Email Delivery Service API
//
// API spec for managing OCI Email Delivery services.
//

package email

import (
	"github.com/oracle/oci-go-sdk/common"
)

// Sender The full information representing an approved sender.
type Sender struct {

	// Email address of the sender.
	EmailAddress *string `mandatory:"false" json:"emailAddress"`

	// The unique OCID of the sender.
	Id *string `mandatory:"false" json:"id"`

	// Value of the SPF field. For more information about SPF, please see
	// SPF Authentication (https://docs.us-phoenix-1.oraclecloud.com/Content/Email/Concepts/emaildeliveryoverview.htm#spf).
	IsSpf *bool `mandatory:"false" json:"isSpf"`

	// The sender's current lifecycle state.
	LifecycleState SenderLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The date and time the approved sender was added in "YYYY-MM-ddThh:mmZ"
	// format with a Z offset, as defined by RFC 3339.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m Sender) String() string {
	return common.PointerString(m)
}

// SenderLifecycleStateEnum Enum with underlying type: string
type SenderLifecycleStateEnum string

// Set of constants representing the allowable values for SenderLifecycleState
const (
	SenderLifecycleStateCreating SenderLifecycleStateEnum = "CREATING"
	SenderLifecycleStateActive   SenderLifecycleStateEnum = "ACTIVE"
	SenderLifecycleStateDeleting SenderLifecycleStateEnum = "DELETING"
	SenderLifecycleStateDeleted  SenderLifecycleStateEnum = "DELETED"
)

var mappingSenderLifecycleState = map[string]SenderLifecycleStateEnum{
	"CREATING": SenderLifecycleStateCreating,
	"ACTIVE":   SenderLifecycleStateActive,
	"DELETING": SenderLifecycleStateDeleting,
	"DELETED":  SenderLifecycleStateDeleted,
}

// GetSenderLifecycleStateEnumValues Enumerates the set of values for SenderLifecycleState
func GetSenderLifecycleStateEnumValues() []SenderLifecycleStateEnum {
	values := make([]SenderLifecycleStateEnum, 0)
	for _, v := range mappingSenderLifecycleState {
		values = append(values, v)
	}
	return values
}
