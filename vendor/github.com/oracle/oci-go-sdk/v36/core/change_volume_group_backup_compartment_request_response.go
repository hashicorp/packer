// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/v36/common"
	"net/http"
)

// ChangeVolumeGroupBackupCompartmentRequest wrapper for the ChangeVolumeGroupBackupCompartment operation
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeVolumeGroupBackupCompartment.go.html to see an example of how to use ChangeVolumeGroupBackupCompartmentRequest.
type ChangeVolumeGroupBackupCompartmentRequest struct {

	// The Oracle Cloud ID (OCID) that uniquely identifies the volume group backup.
	VolumeGroupBackupId *string `mandatory:"true" contributesTo:"path" name:"volumeGroupBackupId"`

	// Request to change the compartment of given volume group backup.
	ChangeVolumeGroupBackupCompartmentDetails `contributesTo:"body"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ChangeVolumeGroupBackupCompartmentRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ChangeVolumeGroupBackupCompartmentRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ChangeVolumeGroupBackupCompartmentRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ChangeVolumeGroupBackupCompartmentResponse wrapper for the ChangeVolumeGroupBackupCompartment operation
type ChangeVolumeGroupBackupCompartmentResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ChangeVolumeGroupBackupCompartmentResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ChangeVolumeGroupBackupCompartmentResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
