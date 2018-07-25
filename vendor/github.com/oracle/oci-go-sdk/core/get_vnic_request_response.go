// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetVnicRequest wrapper for the GetVnic operation
type GetVnicRequest struct {

	// The OCID of the VNIC.
	VnicId *string `mandatory:"true" contributesTo:"path" name:"vnicId"`
}

func (request GetVnicRequest) String() string {
	return common.PointerString(request)
}

// GetVnicResponse wrapper for the GetVnic operation
type GetVnicResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Vnic instance
	Vnic `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetVnicResponse) String() string {
	return common.PointerString(response)
}
