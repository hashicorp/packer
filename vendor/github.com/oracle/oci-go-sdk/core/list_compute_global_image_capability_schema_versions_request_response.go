// Copyright (c) 2016, 2018, 2020, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListComputeGlobalImageCapabilitySchemaVersionsRequest wrapper for the ListComputeGlobalImageCapabilitySchemaVersions operation
type ListComputeGlobalImageCapabilitySchemaVersionsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compute global image capability schema
	ComputeGlobalImageCapabilitySchemaId *string `mandatory:"true" contributesTo:"path" name:"computeGlobalImageCapabilitySchemaId"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListComputeGlobalImageCapabilitySchemaVersionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListComputeGlobalImageCapabilitySchemaVersionsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListComputeGlobalImageCapabilitySchemaVersionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListComputeGlobalImageCapabilitySchemaVersionsResponse wrapper for the ListComputeGlobalImageCapabilitySchemaVersions operation
type ListComputeGlobalImageCapabilitySchemaVersionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ComputeGlobalImageCapabilitySchemaVersionSummary instances
	Items []ComputeGlobalImageCapabilitySchemaVersionSummary `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListComputeGlobalImageCapabilitySchemaVersionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListComputeGlobalImageCapabilitySchemaVersionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum Enum with underlying type: string
type ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum string

// Set of constants representing the allowable values for ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum
const (
	ListComputeGlobalImageCapabilitySchemaVersionsSortByTimecreated ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum = "TIMECREATED"
	ListComputeGlobalImageCapabilitySchemaVersionsSortByDisplayname ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum = "DISPLAYNAME"
)

var mappingListComputeGlobalImageCapabilitySchemaVersionsSortBy = map[string]ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum{
	"TIMECREATED": ListComputeGlobalImageCapabilitySchemaVersionsSortByTimecreated,
	"DISPLAYNAME": ListComputeGlobalImageCapabilitySchemaVersionsSortByDisplayname,
}

// GetListComputeGlobalImageCapabilitySchemaVersionsSortByEnumValues Enumerates the set of values for ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum
func GetListComputeGlobalImageCapabilitySchemaVersionsSortByEnumValues() []ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum {
	values := make([]ListComputeGlobalImageCapabilitySchemaVersionsSortByEnum, 0)
	for _, v := range mappingListComputeGlobalImageCapabilitySchemaVersionsSortBy {
		values = append(values, v)
	}
	return values
}

// ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum Enum with underlying type: string
type ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum string

// Set of constants representing the allowable values for ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum
const (
	ListComputeGlobalImageCapabilitySchemaVersionsSortOrderAsc  ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum = "ASC"
	ListComputeGlobalImageCapabilitySchemaVersionsSortOrderDesc ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum = "DESC"
)

var mappingListComputeGlobalImageCapabilitySchemaVersionsSortOrder = map[string]ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum{
	"ASC":  ListComputeGlobalImageCapabilitySchemaVersionsSortOrderAsc,
	"DESC": ListComputeGlobalImageCapabilitySchemaVersionsSortOrderDesc,
}

// GetListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnumValues Enumerates the set of values for ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum
func GetListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnumValues() []ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum {
	values := make([]ListComputeGlobalImageCapabilitySchemaVersionsSortOrderEnum, 0)
	for _, v := range mappingListComputeGlobalImageCapabilitySchemaVersionsSortOrder {
		values = append(values, v)
	}
	return values
}
