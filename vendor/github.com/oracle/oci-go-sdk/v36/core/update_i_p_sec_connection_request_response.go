// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/v36/common"
	"net/http"
)

// UpdateIPSecConnectionRequest wrapper for the UpdateIPSecConnection operation
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateIPSecConnection.go.html to see an example of how to use UpdateIPSecConnectionRequest.
type UpdateIPSecConnectionRequest struct {

	// The OCID of the IPSec connection.
	IpscId *string `mandatory:"true" contributesTo:"path" name:"ipscId"`

	// Details object for updating a IPSec connection.
	UpdateIpSecConnectionDetails `contributesTo:"body"`

	// For optimistic concurrency control. In the PUT or DELETE call for a resource, set the `if-match`
	// parameter to the value of the etag from a previous GET or POST response for that resource. The resource
	// will be updated or deleted only if the etag you provide matches the resource's current etag value.
	IfMatch *string `mandatory:"false" contributesTo:"header" name:"if-match"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request UpdateIPSecConnectionRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request UpdateIPSecConnectionRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request UpdateIPSecConnectionRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// UpdateIPSecConnectionResponse wrapper for the UpdateIPSecConnection operation
type UpdateIPSecConnectionResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The IpSecConnection instance
	IpSecConnection `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response UpdateIPSecConnectionResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response UpdateIPSecConnectionResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
