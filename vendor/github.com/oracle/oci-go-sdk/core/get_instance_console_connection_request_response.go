// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetInstanceConsoleConnectionRequest wrapper for the GetInstanceConsoleConnection operation
type GetInstanceConsoleConnectionRequest struct {

	// The OCID of the intance console connection
	InstanceConsoleConnectionId *string `mandatory:"true" contributesTo:"path" name:"instanceConsoleConnectionId"`
}

func (request GetInstanceConsoleConnectionRequest) String() string {
	return common.PointerString(request)
}

// GetInstanceConsoleConnectionResponse wrapper for the GetInstanceConsoleConnection operation
type GetInstanceConsoleConnectionResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The InstanceConsoleConnection instance
	InstanceConsoleConnection `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetInstanceConsoleConnectionResponse) String() string {
	return common.PointerString(response)
}
