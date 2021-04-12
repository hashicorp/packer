// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/v36/common"
	"net/http"
)

// DeleteVolumeGroupRequest wrapper for the DeleteVolumeGroup operation
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteVolumeGroup.go.html to see an example of how to use DeleteVolumeGroupRequest.
type DeleteVolumeGroupRequest struct {

	// The Oracle Cloud ID (OCID) that uniquely identifies the volume group.
	VolumeGroupId *string `mandatory:"true" contributesTo:"path" name:"volumeGroupId"`

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

func (request DeleteVolumeGroupRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request DeleteVolumeGroupRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request DeleteVolumeGroupRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// DeleteVolumeGroupResponse wrapper for the DeleteVolumeGroup operation
type DeleteVolumeGroupResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response DeleteVolumeGroupResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response DeleteVolumeGroupResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
