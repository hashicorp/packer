// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package database

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetDatabaseRequest wrapper for the GetDatabase operation
type GetDatabaseRequest struct {

	// The database OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm).
	DatabaseId *string `mandatory:"true" contributesTo:"path" name:"databaseId"`
}

func (request GetDatabaseRequest) String() string {
	return common.PointerString(request)
}

// GetDatabaseResponse wrapper for the GetDatabase operation
type GetDatabaseResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Database instance
	Database `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetDatabaseResponse) String() string {
	return common.PointerString(response)
}
