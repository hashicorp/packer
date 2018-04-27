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

// Bucket To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type Bucket struct {

	// The namespace in which the bucket lives.
	Namespace *string `mandatory:"true" json:"namespace"`

	// The name of the bucket.
	Name *string `mandatory:"true" json:"name"`

	// The compartment ID in which the bucket is authorized.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Arbitrary string keys and values for user-defined metadata.
	Metadata map[string]string `mandatory:"true" json:"metadata"`

	// The OCID of the user who created the bucket.
	CreatedBy *string `mandatory:"true" json:"createdBy"`

	// The date and time at which the bucket was created.
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The entity tag for the bucket.
	Etag *string `mandatory:"true" json:"etag"`

	// The type of public access available on this bucket. Allows authenticated caller to access the bucket or
	// contents of this bucket. By default a bucket is set to NoPublicAccess. It is treated as NoPublicAccess
	// when this value is not specified. When the type is NoPublicAccess the bucket does not allow any public access.
	// When the type is ObjectRead the bucket allows public access to the GetObject, HeadObject, ListObjects.
	PublicAccessType BucketPublicAccessTypeEnum `mandatory:"false" json:"publicAccessType,omitempty"`
}

func (m Bucket) String() string {
	return common.PointerString(m)
}

// BucketPublicAccessTypeEnum Enum with underlying type: string
type BucketPublicAccessTypeEnum string

// Set of constants representing the allowable values for BucketPublicAccessType
const (
	BucketPublicAccessTypeNopublicaccess BucketPublicAccessTypeEnum = "NoPublicAccess"
	BucketPublicAccessTypeObjectread     BucketPublicAccessTypeEnum = "ObjectRead"
)

var mappingBucketPublicAccessType = map[string]BucketPublicAccessTypeEnum{
	"NoPublicAccess": BucketPublicAccessTypeNopublicaccess,
	"ObjectRead":     BucketPublicAccessTypeObjectread,
}

// GetBucketPublicAccessTypeEnumValues Enumerates the set of values for BucketPublicAccessType
func GetBucketPublicAccessTypeEnumValues() []BucketPublicAccessTypeEnum {
	values := make([]BucketPublicAccessTypeEnum, 0)
	for _, v := range mappingBucketPublicAccessType {
		values = append(values, v)
	}
	return values
}
