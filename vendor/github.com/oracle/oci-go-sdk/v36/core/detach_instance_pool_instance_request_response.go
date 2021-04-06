// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/v36/common"
	"net/http"
)

// DetachInstancePoolInstanceRequest wrapper for the DetachInstancePoolInstance operation
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DetachInstancePoolInstance.go.html to see an example of how to use DetachInstancePoolInstanceRequest.
type DetachInstancePoolInstanceRequest struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the instance pool.
	InstancePoolId *string `mandatory:"true" contributesTo:"path" name:"instancePoolId"`

	// Instance being detached
	DetachInstancePoolInstanceDetails `contributesTo:"body"`

	// A token that uniquely identifies a request so it can be retried in case of a timeout or
	// server error without risk of executing that same action again. Retry tokens expire after 24
	// hours, but can be invalidated before then due to conflicting operations (for example, if a resource
	// has been deleted and purged from the system, then a retry of the original creation request
	// may be rejected).
	OpcRetryToken *string `mandatory:"false" contributesTo:"header" name:"opc-retry-token"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request DetachInstancePoolInstanceRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request DetachInstancePoolInstanceRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request DetachInstancePoolInstanceRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// DetachInstancePoolInstanceResponse wrapper for the DetachInstancePoolInstance operation
type DetachInstancePoolInstanceResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// The OCID of the work request. Use GetWorkRequest (https://docs.cloud.oracle.com/api/#/en/workrequests/20160918/WorkRequest/GetWorkRequest)
	// with this ID to track the status of the request.
	OpcWorkRequestId *string `presentIn:"header" name:"opc-work-request-id"`
}

func (response DetachInstancePoolInstanceResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response DetachInstancePoolInstanceResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
