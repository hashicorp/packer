// Copyright (c) 2016, 2018, 2020, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetComputeImageCapabilitySchemaRequest wrapper for the GetComputeImageCapabilitySchema operation
type GetComputeImageCapabilitySchemaRequest struct {

	// The id of the compute image capability schema or the image ocid
	ComputeImageCapabilitySchemaId *string `mandatory:"true" contributesTo:"path" name:"computeImageCapabilitySchemaId"`

	// Merge the image capability schema with the global image capability schema
	IsMergeEnabled *bool `mandatory:"false" contributesTo:"query" name:"isMergeEnabled"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetComputeImageCapabilitySchemaRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetComputeImageCapabilitySchemaRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetComputeImageCapabilitySchemaRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// GetComputeImageCapabilitySchemaResponse wrapper for the GetComputeImageCapabilitySchema operation
type GetComputeImageCapabilitySchemaResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The ComputeImageCapabilitySchema instance
	ComputeImageCapabilitySchema `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetComputeImageCapabilitySchemaResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetComputeImageCapabilitySchemaResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
