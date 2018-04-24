// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetWindowsInstanceInitialCredentialsRequest wrapper for the GetWindowsInstanceInitialCredentials operation
type GetWindowsInstanceInitialCredentialsRequest struct {

	// The OCID of the instance.
	InstanceId *string `mandatory:"true" contributesTo:"path" name:"instanceId"`
}

func (request GetWindowsInstanceInitialCredentialsRequest) String() string {
	return common.PointerString(request)
}

// GetWindowsInstanceInitialCredentialsResponse wrapper for the GetWindowsInstanceInitialCredentials operation
type GetWindowsInstanceInitialCredentialsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The InstanceCredentials instance
	InstanceCredentials `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetWindowsInstanceInitialCredentialsResponse) String() string {
	return common.PointerString(response)
}
