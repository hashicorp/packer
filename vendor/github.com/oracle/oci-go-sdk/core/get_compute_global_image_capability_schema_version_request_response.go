// Copyright (c) 2016, 2018, 2020, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetComputeGlobalImageCapabilitySchemaVersionRequest wrapper for the GetComputeGlobalImageCapabilitySchemaVersion operation
type GetComputeGlobalImageCapabilitySchemaVersionRequest struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compute global image capability schema
	ComputeGlobalImageCapabilitySchemaId *string `mandatory:"true" contributesTo:"path" name:"computeGlobalImageCapabilitySchemaId"`

	// The name of the compute global image capability schema version
	ComputeGlobalImageCapabilitySchemaVersionName *string `mandatory:"true" contributesTo:"path" name:"computeGlobalImageCapabilitySchemaVersionName"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request GetComputeGlobalImageCapabilitySchemaVersionRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request GetComputeGlobalImageCapabilitySchemaVersionRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request GetComputeGlobalImageCapabilitySchemaVersionRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// GetComputeGlobalImageCapabilitySchemaVersionResponse wrapper for the GetComputeGlobalImageCapabilitySchemaVersion operation
type GetComputeGlobalImageCapabilitySchemaVersionResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The ComputeGlobalImageCapabilitySchemaVersion instance
	ComputeGlobalImageCapabilitySchemaVersion `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetComputeGlobalImageCapabilitySchemaVersionResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response GetComputeGlobalImageCapabilitySchemaVersionResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
