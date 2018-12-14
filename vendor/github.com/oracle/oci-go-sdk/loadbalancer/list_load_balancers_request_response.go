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

	// The field to sort by.  You can provide one sort order (`sortOrder`). Default order for TIMECREATED is descending.
	// Default order for DISPLAYNAME is ascending. The DISPLAYNAME sort order is case sensitive.
	SortBy ListLoadBalancersSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending (`ASC`) or descending (`DESC`). The DISPLAYNAME sort order is case sensitive.
	SortOrder ListLoadBalancersSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// A filter to return only resources that match the given display name exactly.
	// Example: `example_load_balancer`
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// A filter to return only resources that match the given lifecycle state.
	// Example: `SUCCEEDED`
	LifecycleState LoadBalancerLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListLoadBalancersRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListLoadBalancersRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListLoadBalancersRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListLoadBalancersResponse wrapper for the ListLoadBalancers operation
type ListLoadBalancersResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []LoadBalancer instances
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

// HTTPResponse implements the OCIResponse interface
func (response ListLoadBalancersResponse) HTTPResponse() *http.Response {
	return response.RawResponse
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
