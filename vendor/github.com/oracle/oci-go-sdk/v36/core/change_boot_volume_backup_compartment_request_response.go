// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/v36/common"
	"net/http"
)

// ChangeBootVolumeBackupCompartmentRequest wrapper for the ChangeBootVolumeBackupCompartment operation
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeBootVolumeBackupCompartment.go.html to see an example of how to use ChangeBootVolumeBackupCompartmentRequest.
type ChangeBootVolumeBackupCompartmentRequest struct {

	// The OCID of the boot volume backup.
	BootVolumeBackupId *string `mandatory:"true" contributesTo:"path" name:"bootVolumeBackupId"`

	// Request to change the compartment of given boot volume backup.
	ChangeBootVolumeBackupCompartmentDetails `contributesTo:"body"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ChangeBootVolumeBackupCompartmentRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ChangeBootVolumeBackupCompartmentRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ChangeBootVolumeBackupCompartmentRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ChangeBootVolumeBackupCompartmentResponse wrapper for the ChangeBootVolumeBackupCompartment operation
type ChangeBootVolumeBackupCompartmentResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ChangeBootVolumeBackupCompartmentResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ChangeBootVolumeBackupCompartmentResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
