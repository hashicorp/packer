// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package database

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListDbVersionsRequest wrapper for the ListDbVersions operation
type ListDbVersionsRequest struct {

	// The compartment OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm).
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The maximum number of items to return.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The pagination token to continue listing from.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// If provided, filters the results to the set of database versions which are supported for the given shape.
	DbSystemShape *string `mandatory:"false" contributesTo:"query" name:"dbSystemShape"`
}

func (request ListDbVersionsRequest) String() string {
	return common.PointerString(request)
}

// ListDbVersionsResponse wrapper for the ListDbVersions operation
type ListDbVersionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []DbVersionSummary instance
	Items []DbVersionSummary `presentIn:"body"`

	// For pagination of a list of items. When paging through a list, if this header appears in the response,
	// then there are additional items still to get. Include this value as the `page` parameter for the
	// subsequent GET request. For information about pagination, see
	// List Pagination (https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/usingapi.htm#List_Pagination).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListDbVersionsResponse) String() string {
	return common.PointerString(response)
}
