// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/v36/common"
	"net/http"
)

// ListVcnsRequest wrapper for the ListVcns operation
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListVcns.go.html to see an example of how to use ListVcnsRequest.
type ListVcnsRequest struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// For list pagination. The maximum number of results per page, or items to return in a paginated
	// "List" call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	// Example: `50`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// For list pagination. The value of the `opc-next-page` response header from the previous "List"
	// call. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// A filter to return only resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The field to sort by. You can provide one sort order (`sortOrder`). Default order for
	// TIMECREATED is descending. Default order for DISPLAYNAME is ascending. The DISPLAYNAME
	// sort order is case sensitive.
	// **Note:** In general, some "List" operations (for example, `ListInstances`) let you
	// optionally filter by availability domain if the scope of the resource type is within a
	// single availability domain. If you call one of these "List" operations without specifying
	// an availability domain, the resources are grouped by availability domain, then sorted.
	SortBy ListVcnsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListVcnsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given lifecycle
	// state. The state value is case-insensitive.
	LifecycleState VcnLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListVcnsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListVcnsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListVcnsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListVcnsResponse wrapper for the ListVcns operation
type ListVcnsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []Vcn instances
	Items []Vcn `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListVcnsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListVcnsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListVcnsSortByEnum Enum with underlying type: string
type ListVcnsSortByEnum string

// Set of constants representing the allowable values for ListVcnsSortByEnum
const (
	ListVcnsSortByTimecreated ListVcnsSortByEnum = "TIMECREATED"
	ListVcnsSortByDisplayname ListVcnsSortByEnum = "DISPLAYNAME"
)

var mappingListVcnsSortBy = map[string]ListVcnsSortByEnum{
	"TIMECREATED": ListVcnsSortByTimecreated,
	"DISPLAYNAME": ListVcnsSortByDisplayname,
}

// GetListVcnsSortByEnumValues Enumerates the set of values for ListVcnsSortByEnum
func GetListVcnsSortByEnumValues() []ListVcnsSortByEnum {
	values := make([]ListVcnsSortByEnum, 0)
	for _, v := range mappingListVcnsSortBy {
		values = append(values, v)
	}
	return values
}

// ListVcnsSortOrderEnum Enum with underlying type: string
type ListVcnsSortOrderEnum string

// Set of constants representing the allowable values for ListVcnsSortOrderEnum
const (
	ListVcnsSortOrderAsc  ListVcnsSortOrderEnum = "ASC"
	ListVcnsSortOrderDesc ListVcnsSortOrderEnum = "DESC"
)

var mappingListVcnsSortOrder = map[string]ListVcnsSortOrderEnum{
	"ASC":  ListVcnsSortOrderAsc,
	"DESC": ListVcnsSortOrderDesc,
}

// GetListVcnsSortOrderEnumValues Enumerates the set of values for ListVcnsSortOrderEnum
func GetListVcnsSortOrderEnumValues() []ListVcnsSortOrderEnum {
	values := make([]ListVcnsSortOrderEnum, 0)
	for _, v := range mappingListVcnsSortOrder {
		values = append(values, v)
	}
	return values
}
