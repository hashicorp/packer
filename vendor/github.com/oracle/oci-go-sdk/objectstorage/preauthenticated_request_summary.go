// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Object Storage Service API
//
// APIs for managing buckets and objects.
//

package objectstorage

import (
	"github.com/oracle/oci-go-sdk/common"
)

// PreauthenticatedRequestSummary The representation of PreauthenticatedRequestSummary
type PreauthenticatedRequestSummary struct {

	// the unique identifier to use when directly addressing the pre-authenticated request
	Id *string `mandatory:"true" json:"id"`

	// the user supplied name of the pre-authenticated request
	Name *string `mandatory:"true" json:"name"`

	// the operation that can be performed on this resource e.g PUT or GET.
	AccessType PreauthenticatedRequestSummaryAccessTypeEnum `mandatory:"true" json:"accessType"`

	// the expiration date after which the pre authenticated request will no longer be valid as per spec
	// RFC 3339 (https://tools.ietf.org/rfc/rfc3339)
	TimeExpires *common.SDKTime `mandatory:"true" json:"timeExpires"`

	// the date when the pre-authenticated request was created as per spec
	// RFC 3339 (https://tools.ietf.org/rfc/rfc3339)
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// Name of object that is being granted access to by the pre-authenticated request. This can be null and that would mean that the pre-authenticated request is granting access to the entire bucket
	ObjectName *string `mandatory:"false" json:"objectName"`
}

func (m PreauthenticatedRequestSummary) String() string {
	return common.PointerString(m)
}

// PreauthenticatedRequestSummaryAccessTypeEnum Enum with underlying type: string
type PreauthenticatedRequestSummaryAccessTypeEnum string

// Set of constants representing the allowable values for PreauthenticatedRequestSummaryAccessType
const (
	PreauthenticatedRequestSummaryAccessTypeObjectread      PreauthenticatedRequestSummaryAccessTypeEnum = "ObjectRead"
	PreauthenticatedRequestSummaryAccessTypeObjectwrite     PreauthenticatedRequestSummaryAccessTypeEnum = "ObjectWrite"
	PreauthenticatedRequestSummaryAccessTypeObjectreadwrite PreauthenticatedRequestSummaryAccessTypeEnum = "ObjectReadWrite"
	PreauthenticatedRequestSummaryAccessTypeAnyobjectwrite  PreauthenticatedRequestSummaryAccessTypeEnum = "AnyObjectWrite"
)

var mappingPreauthenticatedRequestSummaryAccessType = map[string]PreauthenticatedRequestSummaryAccessTypeEnum{
	"ObjectRead":      PreauthenticatedRequestSummaryAccessTypeObjectread,
	"ObjectWrite":     PreauthenticatedRequestSummaryAccessTypeObjectwrite,
	"ObjectReadWrite": PreauthenticatedRequestSummaryAccessTypeObjectreadwrite,
	"AnyObjectWrite":  PreauthenticatedRequestSummaryAccessTypeAnyobjectwrite,
}

// GetPreauthenticatedRequestSummaryAccessTypeEnumValues Enumerates the set of values for PreauthenticatedRequestSummaryAccessType
func GetPreauthenticatedRequestSummaryAccessTypeEnumValues() []PreauthenticatedRequestSummaryAccessTypeEnum {
	values := make([]PreauthenticatedRequestSummaryAccessTypeEnum, 0)
	for _, v := range mappingPreauthenticatedRequestSummaryAccessType {
		values = append(values, v)
	}
	return values
}
