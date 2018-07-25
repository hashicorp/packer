// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package objectstorage

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// CreatePreauthenticatedRequestRequest wrapper for the CreatePreauthenticatedRequest operation
type CreatePreauthenticatedRequestRequest struct {

	// The top-level namespace used for the request.
	NamespaceName *string `mandatory:"true" contributesTo:"path" name:"namespaceName"`

	// The name of the bucket.
	// Example: `my-new-bucket1`
	BucketName *string `mandatory:"true" contributesTo:"path" name:"bucketName"`

	// details for creating the pre-authenticated request.
	CreatePreauthenticatedRequestDetails `contributesTo:"body"`

	// The client request ID for tracing.
	OpcClientRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-client-request-id"`
}

func (request CreatePreauthenticatedRequestRequest) String() string {
	return common.PointerString(request)
}

// CreatePreauthenticatedRequestResponse wrapper for the CreatePreauthenticatedRequest operation
type CreatePreauthenticatedRequestResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The PreauthenticatedRequest instance
	PreauthenticatedRequest `presentIn:"body"`

	// Echoes back the value passed in the opc-client-request-id header, for use by clients when debugging.
	OpcClientRequestId *string `presentIn:"header" name:"opc-client-request-id"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a particular
	// request, please provide this request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response CreatePreauthenticatedRequestResponse) String() string {
	return common.PointerString(response)
}
