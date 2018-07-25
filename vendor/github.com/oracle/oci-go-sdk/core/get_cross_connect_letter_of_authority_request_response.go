// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetCrossConnectLetterOfAuthorityRequest wrapper for the GetCrossConnectLetterOfAuthority operation
type GetCrossConnectLetterOfAuthorityRequest struct {

	// The OCID of the cross-connect.
	CrossConnectId *string `mandatory:"true" contributesTo:"path" name:"crossConnectId"`
}

func (request GetCrossConnectLetterOfAuthorityRequest) String() string {
	return common.PointerString(request)
}

// GetCrossConnectLetterOfAuthorityResponse wrapper for the GetCrossConnectLetterOfAuthority operation
type GetCrossConnectLetterOfAuthorityResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The LetterOfAuthority instance
	LetterOfAuthority `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetCrossConnectLetterOfAuthorityResponse) String() string {
	return common.PointerString(response)
}
