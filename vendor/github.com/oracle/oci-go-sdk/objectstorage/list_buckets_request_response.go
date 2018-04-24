// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package objectstorage

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListBucketsRequest wrapper for the ListBuckets operation
type ListBucketsRequest struct {

	// The top-level namespace used for the request.
	NamespaceName *string `mandatory:"true" contributesTo:"path" name:"namespaceName"`

	// The ID of the compartment in which to create the bucket.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The maximum number of items to return.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The page at which to start retrieving results.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The client request ID for tracing.
	OpcClientRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-client-request-id"`
}

func (request ListBucketsRequest) String() string {
	return common.PointerString(request)
}

// ListBucketsResponse wrapper for the ListBuckets operation
type ListBucketsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []BucketSummary instance
	Items []BucketSummary `presentIn:"body"`

	// Echoes back the value passed in the opc-client-request-id header, for use by clients when debugging.
	OpcClientRequestId *string `presentIn:"header" name:"opc-client-request-id"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a particular
	// request, please provide this request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For pagination of a list of `Bucket`s. If this header appears in the response, then this
	// is a partial list of buckets. Include this value as the `page` parameter in a subsequent
	// GET request to get the next batch of buckets. For information about pagination, see
	// List Pagination (https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListBucketsResponse) String() string {
	return common.PointerString(response)
}
