// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package loadbalancer

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetWorkRequestRequest wrapper for the GetWorkRequest operation
type GetWorkRequestRequest struct {

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm) of the work request to retrieve.
	WorkRequestId *string `mandatory:"true" contributesTo:"path" name:"workRequestId"`

	// The unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`
}

func (request GetWorkRequestRequest) String() string {
	return common.PointerString(request)
}

// GetWorkRequestResponse wrapper for the GetWorkRequest operation
type GetWorkRequestResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The WorkRequest instance
	WorkRequest `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetWorkRequestResponse) String() string {
	return common.PointerString(response)
}
