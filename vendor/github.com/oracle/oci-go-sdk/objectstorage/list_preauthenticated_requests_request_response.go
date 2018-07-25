// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package objectstorage

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListPreauthenticatedRequestsRequest wrapper for the ListPreauthenticatedRequests operation
type ListPreauthenticatedRequestsRequest struct {

	// The top-level namespace used for the request.
	NamespaceName *string `mandatory:"true" contributesTo:"path" name:"namespaceName"`

	// The name of the bucket.
	// Example: `my-new-bucket1`
	BucketName *string `mandatory:"true" contributesTo:"path" name:"bucketName"`

	// Pre-authenticated requests returned by the list must have object names starting with prefix
	ObjectNamePrefix *string `mandatory:"false" contributesTo:"query" name:"objectNamePrefix"`

	// The maximum number of items to return.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The page at which to start retrieving results.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The client request ID for tracing.
	OpcClientRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-client-request-id"`
}

func (request ListPreauthenticatedRequestsRequest) String() string {
	return common.PointerString(request)
}

// ListPreauthenticatedRequestsResponse wrapper for the ListPreauthenticatedRequests operation
type ListPreauthenticatedRequestsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []PreauthenticatedRequestSummary instance
	Items []PreauthenticatedRequestSummary `presentIn:"body"`

	// Echoes back the value passed in the opc-client-request-id header, for use by clients when debugging.
	OpcClientRequestId *string `presentIn:"header" name:"opc-client-request-id"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a particular
	// request, please provide this request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For pagination of a list of pre-authenticated requests, if this header appears in the response,
	// then this is a partial list. Include this value as the `page` parameter in a subsequent
	// GET request to get the next batch of pre-authenticated requests.
	// For information about pagination, see List Pagination (https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListPreauthenticatedRequestsResponse) String() string {
	return common.PointerString(response)
}
