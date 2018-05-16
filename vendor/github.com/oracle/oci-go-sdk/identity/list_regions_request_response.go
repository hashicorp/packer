// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package identity

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListRegionsRequest wrapper for the ListRegions operation
type ListRegionsRequest struct {
}

func (request ListRegionsRequest) String() string {
	return common.PointerString(request)
}

// ListRegionsResponse wrapper for the ListRegions operation
type ListRegionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []Region instance
	Items []Region `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListRegionsResponse) String() string {
	return common.PointerString(response)
}
