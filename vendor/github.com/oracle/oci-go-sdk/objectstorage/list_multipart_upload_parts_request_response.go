// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package objectstorage

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListMultipartUploadPartsRequest wrapper for the ListMultipartUploadParts operation
type ListMultipartUploadPartsRequest struct {

	// The top-level namespace used for the request.
	NamespaceName *string `mandatory:"true" contributesTo:"path" name:"namespaceName"`

	// The name of the bucket.
	// Example: `my-new-bucket1`
	BucketName *string `mandatory:"true" contributesTo:"path" name:"bucketName"`

	// The name of the object.
	// Example: `test/object1.log`
	ObjectName *string `mandatory:"true" contributesTo:"path" name:"objectName"`

	// The upload ID for a multipart upload.
	UploadId *string `mandatory:"true" contributesTo:"query" name:"uploadId"`

	// The maximum number of items to return.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The page at which to start retrieving results.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The client request ID for tracing.
	OpcClientRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-client-request-id"`
}

func (request ListMultipartUploadPartsRequest) String() string {
	return common.PointerString(request)
}

// ListMultipartUploadPartsResponse wrapper for the ListMultipartUploadParts operation
type ListMultipartUploadPartsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []MultipartUploadPartSummary instance
	Items []MultipartUploadPartSummary `presentIn:"body"`

	// Echoes back the value passed in the opc-client-request-id header, for use by clients when debugging.
	OpcClientRequestId *string `presentIn:"header" name:"opc-client-request-id"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a particular
	// request, please provide this request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For pagination of a list of `MultipartUploadPartSummary`s. If this header appears in the response,
	// then this is a partial list of object parts. Include this value as the `page` parameter in a subsequent
	// GET request to get the next batch of object parts. For information about pagination, see
	// List Pagination (https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/usingapi.htm).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListMultipartUploadPartsResponse) String() string {
	return common.PointerString(response)
}
