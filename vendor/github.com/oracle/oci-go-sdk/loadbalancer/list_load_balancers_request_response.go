// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package loadbalancer

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListLoadBalancersRequest wrapper for the ListLoadBalancers operation
type ListLoadBalancersRequest struct {

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm) of the compartment containing the load balancers to list.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// The maximum number of items to return in a paginated "List" call.
	// Example: `500`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	// Example: `3`
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The level of detail to return for each result. Can be `full` or `simple`.
	// Example: `full`
	Detail *string `mandatory:"false" contributesTo:"query" name:"detail"`

	// The field to sort by.  Only one sort order may be provided.  Time created is default ordered as descending.  Display name is default ordered as ascending.
	SortBy ListLoadBalancersSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either 'asc' or 'desc'
	SortOrder ListLoadBalancersSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to only return resources that match the given display name exactly.
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to only return resources that match the given lifecycle state.
	LifecycleState LoadBalancerLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`
}

func (request ListLoadBalancersRequest) String() string {
	return common.PointerString(request)
}

// ListLoadBalancersResponse wrapper for the ListLoadBalancers operation
type ListLoadBalancersResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []LoadBalancer instance
	Items []LoadBalancer `presentIn:"body"`

	// For pagination of a list of items. When paging through a list, if this header appears in the response,
	// then a partial list might have been returned. Include this value as the `page` parameter for the
	// subsequent GET request to get the next batch of items.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListLoadBalancersResponse) String() string {
	return common.PointerString(response)
}

// ListLoadBalancersSortByEnum Enum with underlying type: string
type ListLoadBalancersSortByEnum string

// Set of constants representing the allowable values for ListLoadBalancersSortBy
const (
	ListLoadBalancersSortByTimecreated ListLoadBalancersSortByEnum = "TIMECREATED"
	ListLoadBalancersSortByDisplayname ListLoadBalancersSortByEnum = "DISPLAYNAME"
)

var mappingListLoadBalancersSortBy = map[string]ListLoadBalancersSortByEnum{
	"TIMECREATED": ListLoadBalancersSortByTimecreated,
	"DISPLAYNAME": ListLoadBalancersSortByDisplayname,
}

// GetListLoadBalancersSortByEnumValues Enumerates the set of values for ListLoadBalancersSortBy
func GetListLoadBalancersSortByEnumValues() []ListLoadBalancersSortByEnum {
	values := make([]ListLoadBalancersSortByEnum, 0)
	for _, v := range mappingListLoadBalancersSortBy {
		values = append(values, v)
	}
	return values
}

// ListLoadBalancersSortOrderEnum Enum with underlying type: string
type ListLoadBalancersSortOrderEnum string

// Set of constants representing the allowable values for ListLoadBalancersSortOrder
const (
	ListLoadBalancersSortOrderAsc  ListLoadBalancersSortOrderEnum = "ASC"
	ListLoadBalancersSortOrderDesc ListLoadBalancersSortOrderEnum = "DESC"
)

var mappingListLoadBalancersSortOrder = map[string]ListLoadBalancersSortOrderEnum{
	"ASC":  ListLoadBalancersSortOrderAsc,
	"DESC": ListLoadBalancersSortOrderDesc,
}

// GetListLoadBalancersSortOrderEnumValues Enumerates the set of values for ListLoadBalancersSortOrder
func GetListLoadBalancersSortOrderEnumValues() []ListLoadBalancersSortOrderEnum {
	values := make([]ListLoadBalancersSortOrderEnum, 0)
	for _, v := range mappingListLoadBalancersSortOrder {
		values = append(values, v)
	}
	return values
}
