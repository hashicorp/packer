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

// Suppression The full information representing an email suppression.
type Suppression struct {

	// Email address of the suppression.
	EmailAddress *string `mandatory:"false" json:"emailAddress"`

	// The unique ID of the suppression.
	Id *string `mandatory:"false" json:"id"`

	// The reason that the email address was suppressed. For more information on the types of bounces, see Suppresion List (https://docs.us-phoenix-1.oraclecloud.com/Content/Email/Concepts/emaildeliveryoverview.htm#suppressionlist).
	Reason SuppressionReasonEnum `mandatory:"false" json:"reason,omitempty"`

	// The date and time the approved sender was added in "YYYY-MM-ddThh:mmZ"
	// format with a Z offset, as defined by RFC 3339.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m Suppression) String() string {
	return common.PointerString(m)
}

// SuppressionReasonEnum Enum with underlying type: string
type SuppressionReasonEnum string

// Set of constants representing the allowable values for SuppressionReason
const (
	SuppressionReasonUnknown     SuppressionReasonEnum = "UNKNOWN"
	SuppressionReasonHardbounce  SuppressionReasonEnum = "HARDBOUNCE"
	SuppressionReasonComplaint   SuppressionReasonEnum = "COMPLAINT"
	SuppressionReasonManual      SuppressionReasonEnum = "MANUAL"
	SuppressionReasonSoftbounce  SuppressionReasonEnum = "SOFTBOUNCE"
	SuppressionReasonUnsubscribe SuppressionReasonEnum = "UNSUBSCRIBE"
)

var mappingSuppressionReason = map[string]SuppressionReasonEnum{
	"UNKNOWN":     SuppressionReasonUnknown,
	"HARDBOUNCE":  SuppressionReasonHardbounce,
	"COMPLAINT":   SuppressionReasonComplaint,
	"MANUAL":      SuppressionReasonManual,
	"SOFTBOUNCE":  SuppressionReasonSoftbounce,
	"UNSUBSCRIBE": SuppressionReasonUnsubscribe,
}

// GetSuppressionReasonEnumValues Enumerates the set of values for SuppressionReason
func GetSuppressionReasonEnumValues() []SuppressionReasonEnum {
	values := make([]SuppressionReasonEnum, 0)
	for _, v := range mappingSuppressionReason {
		values = append(values, v)
	}
	return values
}
