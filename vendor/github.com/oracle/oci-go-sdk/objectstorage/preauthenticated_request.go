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

// PreauthenticatedRequest The representation of PreauthenticatedRequest
type PreauthenticatedRequest struct {

	// the unique identifier to use when directly addressing the pre-authenticated request
	Id *string `mandatory:"true" json:"id"`

	// the user supplied name of the pre-authenticated request.
	Name *string `mandatory:"true" json:"name"`

	// the uri to embed in the url when using the pre-authenticated request.
	AccessUri *string `mandatory:"true" json:"accessUri"`

	// the operation that can be performed on this resource e.g PUT or GET.
	AccessType PreauthenticatedRequestAccessTypeEnum `mandatory:"true" json:"accessType"`

	// the expiration date after which the pre authenticated request will no longer be valid as per spec
	// RFC 3339 (https://tools.ietf.org/rfc/rfc3339)
	TimeExpires *common.SDKTime `mandatory:"true" json:"timeExpires"`

	// the date when the pre-authenticated request was created as per spec
	// RFC 3339 (https://tools.ietf.org/rfc/rfc3339)
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// Name of object that is being granted access to by the pre-authenticated request. This can be null and that would mean that the pre-authenticated request is granting access to the entire bucket
	ObjectName *string `mandatory:"false" json:"objectName"`
}

func (m PreauthenticatedRequest) String() string {
	return common.PointerString(m)
}

// PreauthenticatedRequestAccessTypeEnum Enum with underlying type: string
type PreauthenticatedRequestAccessTypeEnum string

// Set of constants representing the allowable values for PreauthenticatedRequestAccessType
const (
	PreauthenticatedRequestAccessTypeObjectread      PreauthenticatedRequestAccessTypeEnum = "ObjectRead"
	PreauthenticatedRequestAccessTypeObjectwrite     PreauthenticatedRequestAccessTypeEnum = "ObjectWrite"
	PreauthenticatedRequestAccessTypeObjectreadwrite PreauthenticatedRequestAccessTypeEnum = "ObjectReadWrite"
	PreauthenticatedRequestAccessTypeAnyobjectwrite  PreauthenticatedRequestAccessTypeEnum = "AnyObjectWrite"
)

var mappingPreauthenticatedRequestAccessType = map[string]PreauthenticatedRequestAccessTypeEnum{
	"ObjectRead":      PreauthenticatedRequestAccessTypeObjectread,
	"ObjectWrite":     PreauthenticatedRequestAccessTypeObjectwrite,
	"ObjectReadWrite": PreauthenticatedRequestAccessTypeObjectreadwrite,
	"AnyObjectWrite":  PreauthenticatedRequestAccessTypeAnyobjectwrite,
}

// GetPreauthenticatedRequestAccessTypeEnumValues Enumerates the set of values for PreauthenticatedRequestAccessType
func GetPreauthenticatedRequestAccessTypeEnumValues() []PreauthenticatedRequestAccessTypeEnum {
	values := make([]PreauthenticatedRequestAccessTypeEnum, 0)
	for _, v := range mappingPreauthenticatedRequestAccessType {
		values = append(values, v)
	}
	return values
}
