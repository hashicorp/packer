// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/v36/common"
	"net/http"
)

// ConnectRemotePeeringConnectionsRequest wrapper for the ConnectRemotePeeringConnections operation
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ConnectRemotePeeringConnections.go.html to see an example of how to use ConnectRemotePeeringConnectionsRequest.
type ConnectRemotePeeringConnectionsRequest struct {

	// The OCID of the remote peering connection (RPC).
	RemotePeeringConnectionId *string `mandatory:"true" contributesTo:"path" name:"remotePeeringConnectionId"`

	// Details to connect peering connection with peering connection from remote region
	ConnectRemotePeeringConnectionsDetails `contributesTo:"body"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ConnectRemotePeeringConnectionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ConnectRemotePeeringConnectionsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ConnectRemotePeeringConnectionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ConnectRemotePeeringConnectionsResponse wrapper for the ConnectRemotePeeringConnections operation
type ConnectRemotePeeringConnectionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ConnectRemotePeeringConnectionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ConnectRemotePeeringConnectionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
