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

// MultipartUpload To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type MultipartUpload struct {

	// The namespace in which the in-progress multipart upload is stored.
	Namespace *string `mandatory:"true" json:"namespace"`

	// The bucket in which the in-progress multipart upload is stored.
	Bucket *string `mandatory:"true" json:"bucket"`

	// The object name of the in-progress multipart upload.
	Object *string `mandatory:"true" json:"object"`

	// The unique identifier for the in-progress multipart upload.
	UploadId *string `mandatory:"true" json:"uploadId"`

	// The date and time when the upload was created.
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`
}

func (m MultipartUpload) String() string {
	return common.PointerString(m)
}
