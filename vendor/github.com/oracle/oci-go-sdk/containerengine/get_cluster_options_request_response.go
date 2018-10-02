// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package containerengine

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetClusterOptionsRequest wrapper for the GetClusterOptions operation
type GetClusterOptionsRequest struct {

	// The id of the option set to retrieve. Only "all" is supported.
	ClusterOptionId *string `mandatory:"true" contributesTo:"path" name:"clusterOptionId"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetClusterOptionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetClusterOptionsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetClusterOptionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// GetClusterOptionsResponse wrapper for the GetClusterOptions operation
type GetClusterOptionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The ClusterOptions instance
	ClusterOptions `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetClusterOptionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetClusterOptionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
