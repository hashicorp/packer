// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package objectstorage

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// DeletePreauthenticatedRequestRequest wrapper for the DeletePreauthenticatedRequest operation
type DeletePreauthenticatedRequestRequest struct {

	// The top-level namespace used for the request.
	NamespaceName *string `mandatory:"true" contributesTo:"path" name:"namespaceName"`

	// The name of the bucket.
	// Example: `my-new-bucket1`
	BucketName *string `mandatory:"true" contributesTo:"path" name:"bucketName"`

	// The unique identifier for the pre-authenticated request (PAR). This can be used to manage the PAR
	// such as GET or DELETE the PAR
	ParId *string `mandatory:"true" contributesTo:"path" name:"parId"`

	// The client request ID for tracing.
	OpcClientRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-client-request-id"`
}

func (request DeletePreauthenticatedRequestRequest) String() string {
	return common.PointerString(request)
}

// DeletePreauthenticatedRequestResponse wrapper for the DeletePreauthenticatedRequest operation
type DeletePreauthenticatedRequestResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// Echoes back the value passed in the opc-client-request-id header, for use by clients when debugging.
	OpcClientRequestId *string `presentIn:"header" name:"opc-client-request-id"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a particular
	// request, please provide this request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response DeletePreauthenticatedRequestResponse) String() string {
	return common.PointerString(response)
}
