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

// SenderSummary The email addresses and `senderId` representing an approved sender.
type SenderSummary struct {

	// The email address of the sender.
	EmailAddress *string `mandatory:"false" json:"emailAddress"`

	// The unique ID of the sender.
	Id *string `mandatory:"false" json:"id"`

	// The current status of the approved sender.
	LifecycleState SenderSummaryLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// Date time the approved sender was added, in "YYYY-MM-ddThh:mmZ"
	// format with a Z offset, as defined by RFC 3339.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m SenderSummary) String() string {
	return common.PointerString(m)
}

// SenderSummaryLifecycleStateEnum Enum with underlying type: string
type SenderSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for SenderSummaryLifecycleState
const (
	SenderSummaryLifecycleStateCreating SenderSummaryLifecycleStateEnum = "CREATING"
	SenderSummaryLifecycleStateActive   SenderSummaryLifecycleStateEnum = "ACTIVE"
	SenderSummaryLifecycleStateDeleting SenderSummaryLifecycleStateEnum = "DELETING"
	SenderSummaryLifecycleStateDeleted  SenderSummaryLifecycleStateEnum = "DELETED"
)

var mappingSenderSummaryLifecycleState = map[string]SenderSummaryLifecycleStateEnum{
	"CREATING": SenderSummaryLifecycleStateCreating,
	"ACTIVE":   SenderSummaryLifecycleStateActive,
	"DELETING": SenderSummaryLifecycleStateDeleting,
	"DELETED":  SenderSummaryLifecycleStateDeleted,
}

// GetSenderSummaryLifecycleStateEnumValues Enumerates the set of values for SenderSummaryLifecycleState
func GetSenderSummaryLifecycleStateEnumValues() []SenderSummaryLifecycleStateEnum {
	values := make([]SenderSummaryLifecycleStateEnum, 0)
	for _, v := range mappingSenderSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}
