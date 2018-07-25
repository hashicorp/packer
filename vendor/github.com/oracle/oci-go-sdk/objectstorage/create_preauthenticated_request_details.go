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

// CreatePreauthenticatedRequestDetails The representation of CreatePreauthenticatedRequestDetails
type CreatePreauthenticatedRequestDetails struct {

	// user specified name for pre-authenticated request. Helpful for management purposes.
	Name *string `mandatory:"true" json:"name"`

	// the operation that can be performed on this resource e.g PUT or GET.
	AccessType CreatePreauthenticatedRequestDetailsAccessTypeEnum `mandatory:"true" json:"accessType"`

	// The expiration date after which the pre-authenticated request will no longer be valid per spec
	// RFC 3339 (https://tools.ietf.org/rfc/rfc3339)
	TimeExpires *common.SDKTime `mandatory:"true" json:"timeExpires"`

	// Name of object that is being granted access to by the pre-authenticated request. This can be null and that would mean that the pre-authenticated request is granting access to the entire bucket
	ObjectName *string `mandatory:"false" json:"objectName"`
}

func (m CreatePreauthenticatedRequestDetails) String() string {
	return common.PointerString(m)
}

// CreatePreauthenticatedRequestDetailsAccessTypeEnum Enum with underlying type: string
type CreatePreauthenticatedRequestDetailsAccessTypeEnum string

// Set of constants representing the allowable values for CreatePreauthenticatedRequestDetailsAccessType
const (
	CreatePreauthenticatedRequestDetailsAccessTypeObjectread      CreatePreauthenticatedRequestDetailsAccessTypeEnum = "ObjectRead"
	CreatePreauthenticatedRequestDetailsAccessTypeObjectwrite     CreatePreauthenticatedRequestDetailsAccessTypeEnum = "ObjectWrite"
	CreatePreauthenticatedRequestDetailsAccessTypeObjectreadwrite CreatePreauthenticatedRequestDetailsAccessTypeEnum = "ObjectReadWrite"
	CreatePreauthenticatedRequestDetailsAccessTypeAnyobjectwrite  CreatePreauthenticatedRequestDetailsAccessTypeEnum = "AnyObjectWrite"
)

var mappingCreatePreauthenticatedRequestDetailsAccessType = map[string]CreatePreauthenticatedRequestDetailsAccessTypeEnum{
	"ObjectRead":      CreatePreauthenticatedRequestDetailsAccessTypeObjectread,
	"ObjectWrite":     CreatePreauthenticatedRequestDetailsAccessTypeObjectwrite,
	"ObjectReadWrite": CreatePreauthenticatedRequestDetailsAccessTypeObjectreadwrite,
	"AnyObjectWrite":  CreatePreauthenticatedRequestDetailsAccessTypeAnyobjectwrite,
}

// GetCreatePreauthenticatedRequestDetailsAccessTypeEnumValues Enumerates the set of values for CreatePreauthenticatedRequestDetailsAccessType
func GetCreatePreauthenticatedRequestDetailsAccessTypeEnumValues() []CreatePreauthenticatedRequestDetailsAccessTypeEnum {
	values := make([]CreatePreauthenticatedRequestDetailsAccessTypeEnum, 0)
	for _, v := range mappingCreatePreauthenticatedRequestDetailsAccessType {
		values = append(values, v)
	}
	return values
}
