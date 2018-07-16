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

// SuppressionSummary The full information representing a suppression.
type SuppressionSummary struct {

	// The email address of the suppression.
	EmailAddress *string `mandatory:"false" json:"emailAddress"`

	// The unique OCID of the suppression.
	Id *string `mandatory:"false" json:"id"`

	// The reason that the email address was suppressed.
	Reason SuppressionSummaryReasonEnum `mandatory:"false" json:"reason,omitempty"`

	// The date and time a recipient's email address was added to the
	// suppression list, in "YYYY-MM-ddThh:mmZ" format with a Z offset, as
	// defined by RFC 3339.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m SuppressionSummary) String() string {
	return common.PointerString(m)
}

// SuppressionSummaryReasonEnum Enum with underlying type: string
type SuppressionSummaryReasonEnum string

// Set of constants representing the allowable values for SuppressionSummaryReason
const (
	SuppressionSummaryReasonUnknown     SuppressionSummaryReasonEnum = "UNKNOWN"
	SuppressionSummaryReasonHardbounce  SuppressionSummaryReasonEnum = "HARDBOUNCE"
	SuppressionSummaryReasonComplaint   SuppressionSummaryReasonEnum = "COMPLAINT"
	SuppressionSummaryReasonManual      SuppressionSummaryReasonEnum = "MANUAL"
	SuppressionSummaryReasonSoftbounce  SuppressionSummaryReasonEnum = "SOFTBOUNCE"
	SuppressionSummaryReasonUnsubscribe SuppressionSummaryReasonEnum = "UNSUBSCRIBE"
)

var mappingSuppressionSummaryReason = map[string]SuppressionSummaryReasonEnum{
	"UNKNOWN":     SuppressionSummaryReasonUnknown,
	"HARDBOUNCE":  SuppressionSummaryReasonHardbounce,
	"COMPLAINT":   SuppressionSummaryReasonComplaint,
	"MANUAL":      SuppressionSummaryReasonManual,
	"SOFTBOUNCE":  SuppressionSummaryReasonSoftbounce,
	"UNSUBSCRIBE": SuppressionSummaryReasonUnsubscribe,
}

// GetSuppressionSummaryReasonEnumValues Enumerates the set of values for SuppressionSummaryReason
func GetSuppressionSummaryReasonEnumValues() []SuppressionSummaryReasonEnum {
	values := make([]SuppressionSummaryReasonEnum, 0)
	for _, v := range mappingSuppressionSummaryReason {
		values = append(values, v)
	}
	return values
}
