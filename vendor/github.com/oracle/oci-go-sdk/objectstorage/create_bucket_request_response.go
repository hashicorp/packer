// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package objectstorage

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// CreateBucketRequest wrapper for the CreateBucket operation
type CreateBucketRequest struct {

	// The top-level namespace used for the request.
	NamespaceName *string `mandatory:"true" contributesTo:"path" name:"namespaceName"`

	// Request object for creating a bucket.
	CreateBucketDetails `contributesTo:"body"`

	// The client request ID for tracing.
	OpcClientRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-client-request-id"`
}

func (request CreateBucketRequest) String() string {
	return common.PointerString(request)
}

// CreateBucketResponse wrapper for the CreateBucket operation
type CreateBucketResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Bucket instance
	Bucket `presentIn:"body"`

	// Echoes back the value passed in the opc-client-request-id header, for use by clients when debugging.
	OpcClientRequestId *string `presentIn:"header" name:"opc-client-request-id"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a particular
	// request, please provide this request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// The entity tag for the bucket that was created.
	ETag *string `presentIn:"header" name:"etag"`

	// The full path to the bucket that was created.
	Location *string `presentIn:"header" name:"location"`
}

func (response CreateBucketResponse) String() string {
	return common.PointerString(response)
}
