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

// ObjectSummary To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type ObjectSummary struct {

	// The name of the object.
	Name *string `mandatory:"true" json:"name"`

	// Size of the object in bytes.
	Size *int `mandatory:"false" json:"size"`

	// Base64-encoded MD5 hash of the object data.
	Md5 *string `mandatory:"false" json:"md5"`

	// Date and time of object creation.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m ObjectSummary) String() string {
	return common.PointerString(m)
}
