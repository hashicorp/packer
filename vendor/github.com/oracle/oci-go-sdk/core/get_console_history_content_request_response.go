// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetConsoleHistoryContentRequest wrapper for the GetConsoleHistoryContent operation
type GetConsoleHistoryContentRequest struct {

	// The OCID of the console history.
	InstanceConsoleHistoryId *string `mandatory:"true" contributesTo:"path" name:"instanceConsoleHistoryId"`

	// Offset of the snapshot data to retrieve.
	Offset *int `mandatory:"false" contributesTo:"query" name:"offset"`

	// Length of the snapshot data to retrieve.
	Length *int `mandatory:"false" contributesTo:"query" name:"length"`
}

func (request GetConsoleHistoryContentRequest) String() string {
	return common.PointerString(request)
}

// GetConsoleHistoryContentResponse wrapper for the GetConsoleHistoryContent operation
type GetConsoleHistoryContentResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The string instance
	Value *string `presentIn:"body" encoding:"plain-text"`

	// The number of bytes remaining in the snapshot.
	OpcBytesRemaining *int `presentIn:"header" name:"opc-bytes-remaining"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetConsoleHistoryContentResponse) String() string {
	return common.PointerString(response)
}
