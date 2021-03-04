// Copyright (c) 2016, 2018, 2020, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListVlansRequest wrapper for the ListVlans operation
type ListVlansRequest struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the VCN.
	VcnId *string `mandatory:"true" contributesTo:"query" name:"vcnId"`

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
	SortBy ListVlansSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order
	// is case sensitive.
	SortOrder ListVlansSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// A filter to only return resources that match the given lifecycle state.  The state value is case-insensitive.
	LifecycleState VlanLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListVlansRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListVlansRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListVlansRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListVlansResponse wrapper for the ListVlans operation
type ListVlansResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []Vlan instances
	Items []Vlan `presentIn:"body"`

	// For list pagination. When this header appears in the response, additional pages
	// of results remain. For important details about how pagination works, see
	// List Pagination (https://docs.cloud.oracle.com/iaas/Content/API/Concepts/usingapi.htm#nine).
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListVlansResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListVlansResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListVlansSortByEnum Enum with underlying type: string
type ListVlansSortByEnum string

// Set of constants representing the allowable values for ListVlansSortByEnum
const (
	ListVlansSortByTimecreated ListVlansSortByEnum = "TIMECREATED"
	ListVlansSortByDisplayname ListVlansSortByEnum = "DISPLAYNAME"
)

var mappingListVlansSortBy = map[string]ListVlansSortByEnum{
	"TIMECREATED": ListVlansSortByTimecreated,
	"DISPLAYNAME": ListVlansSortByDisplayname,
}

// GetListVlansSortByEnumValues Enumerates the set of values for ListVlansSortByEnum
func GetListVlansSortByEnumValues() []ListVlansSortByEnum {
	values := make([]ListVlansSortByEnum, 0)
	for _, v := range mappingListVlansSortBy {
		values = append(values, v)
	}
	return values
}

// ListVlansSortOrderEnum Enum with underlying type: string
type ListVlansSortOrderEnum string

// Set of constants representing the allowable values for ListVlansSortOrderEnum
const (
	ListVlansSortOrderAsc  ListVlansSortOrderEnum = "ASC"
	ListVlansSortOrderDesc ListVlansSortOrderEnum = "DESC"
)

var mappingListVlansSortOrder = map[string]ListVlansSortOrderEnum{
	"ASC":  ListVlansSortOrderAsc,
	"DESC": ListVlansSortOrderDesc,
}

// GetListVlansSortOrderEnumValues Enumerates the set of values for ListVlansSortOrderEnum
func GetListVlansSortOrderEnumValues() []ListVlansSortOrderEnum {
	values := make([]ListVlansSortOrderEnum, 0)
	for _, v := range mappingListVlansSortOrder {
		values = append(values, v)
	}
	return values
}
