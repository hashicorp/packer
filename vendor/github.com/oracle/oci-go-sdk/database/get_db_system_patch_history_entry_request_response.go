// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package database

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetDbSystemPatchHistoryEntryRequest wrapper for the GetDbSystemPatchHistoryEntry operation
type GetDbSystemPatchHistoryEntryRequest struct {

	// The DB System OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm).
	DbSystemId *string `mandatory:"true" contributesTo:"path" name:"dbSystemId"`

	// The OCID of the patch history entry.
	PatchHistoryEntryId *string `mandatory:"true" contributesTo:"path" name:"patchHistoryEntryId"`
}

func (request GetDbSystemPatchHistoryEntryRequest) String() string {
	return common.PointerString(request)
}

// GetDbSystemPatchHistoryEntryResponse wrapper for the GetDbSystemPatchHistoryEntry operation
type GetDbSystemPatchHistoryEntryResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The PatchHistoryEntry instance
	PatchHistoryEntry `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetDbSystemPatchHistoryEntryResponse) String() string {
	return common.PointerString(response)
}
