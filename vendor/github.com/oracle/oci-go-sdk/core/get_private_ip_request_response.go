// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetPrivateIpRequest wrapper for the GetPrivateIp operation
type GetPrivateIpRequest struct {

	// The private IP's OCID.
	PrivateIpId *string `mandatory:"true" contributesTo:"path" name:"privateIpId"`
}

func (request GetPrivateIpRequest) String() string {
	return common.PointerString(request)
}

// GetPrivateIpResponse wrapper for the GetPrivateIp operation
type GetPrivateIpResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The PrivateIp instance
	PrivateIp `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetPrivateIpResponse) String() string {
	return common.PointerString(response)
}
