// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package database

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetDbSystemRequest wrapper for the GetDbSystem operation
type GetDbSystemRequest struct {

	// The DB System OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm).
	DbSystemId *string `mandatory:"true" contributesTo:"path" name:"dbSystemId"`
}

func (request GetDbSystemRequest) String() string {
	return common.PointerString(request)
}

// GetDbSystemResponse wrapper for the GetDbSystem operation
type GetDbSystemResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The DbSystem instance
	DbSystem `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetDbSystemResponse) String() string {
	return common.PointerString(response)
}
