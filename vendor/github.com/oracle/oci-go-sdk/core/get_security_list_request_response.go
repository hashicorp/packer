// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetSecurityListRequest wrapper for the GetSecurityList operation
type GetSecurityListRequest struct {

	// The OCID of the security list.
	SecurityListId *string `mandatory:"true" contributesTo:"path" name:"securityListId"`
}

func (request GetSecurityListRequest) String() string {
	return common.PointerString(request)
}

// GetSecurityListResponse wrapper for the GetSecurityList operation
type GetSecurityListResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The SecurityList instance
	SecurityList `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetSecurityListResponse) String() string {
	return common.PointerString(response)
}
