// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package database

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListDbSystemPatchHistoryEntriesRequest wrapper for the ListDbSystemPatchHistoryEntries operation
type ListDbSystemPatchHistoryEntriesRequest struct {

	// The DB System OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm).
	DbSystemId *string `mandatory:"true" contributesTo:"path" name:"dbSystemId"`

	// The maximum number of items to return.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The pagination token to continue listing from.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`
}

func (request ListDbSystemPatchHistoryEntriesRequest) String() string {
	return common.PointerString(request)
}

// ListDbSystemPatchHistoryEntriesResponse wrapper for the ListDbSystemPatchHistoryEntries operation
type ListDbSystemPatchHistoryEntriesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []PatchHistoryEntrySummary instance
	Items []PatchHistoryEntrySummary `presentIn:"body"`

	// For pagination of a list of items. When paging through a list, if this header appears in the response,
	// then there are additional items still to get. Include this value as the `page` parameter for the
	// subsequent GET request. For information about pagination, see
	// List Pagination (https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/usingapi.htm#List_Pagination).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListDbSystemPatchHistoryEntriesResponse) String() string {
	return common.PointerString(response)
}
