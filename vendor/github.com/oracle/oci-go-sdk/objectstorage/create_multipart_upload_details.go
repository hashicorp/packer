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

// CreateMultipartUploadDetails To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type CreateMultipartUploadDetails struct {

	// the name of the object to which this multi-part upload is targetted.
	Object *string `mandatory:"true" json:"object"`

	// the content type of the object to upload.
	ContentType *string `mandatory:"false" json:"contentType"`

	// the content language of the object to upload.
	ContentLanguage *string `mandatory:"false" json:"contentLanguage"`

	// the content encoding of the object to upload.
	ContentEncoding *string `mandatory:"false" json:"contentEncoding"`

	// Arbitrary string keys and values for the user-defined metadata for the object.
	// Keys must be in "opc-meta-*" format.
	Metadata map[string]string `mandatory:"false" json:"metadata"`
}

func (m CreateMultipartUploadDetails) String() string {
	return common.PointerString(m)
}
