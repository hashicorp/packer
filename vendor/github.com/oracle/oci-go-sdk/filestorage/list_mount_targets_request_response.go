// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package filestorage

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListMountTargetsRequest wrapper for the ListMountTargets operation
type ListMountTargetsRequest struct {

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The name of the availability domain.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" contributesTo:"query" name:"availabilityDomain"`

	// The maximum number of items to return in a paginated "List" call.
	// Example: `500`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// A user-friendly name. It does not have to be unique, and it is changeable.
	// Example: `My resource`
	DisplayName *string `mandatory:"false" contributesTo:"query" name:"displayName"`

	// The OCID of the export set.
	ExportSetId *string `mandatory:"false" contributesTo:"query" name:"exportSetId"`

	// Filter results by the specified lifecycle state. Must be a valid
	// state for the resource type.
	LifecycleState ListMountTargetsLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Filter results by OCID. Must be an OCID of the correct type for
	// the resouce type.
	Id *string `mandatory:"false" contributesTo:"query" name:"id"`

	// The field to sort by. You can choose either value, but not both.
	// By default, when you sort by time created, results are shown
	// in descending order. When you sort by display name, results are
	// shown in ascending order.
	SortBy ListMountTargetsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either 'asc' or 'desc', where 'asc' is
	// ascending and 'desc' is descending.
	SortOrder ListMountTargetsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListMountTargetsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListMountTargetsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListMountTargetsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListMountTargetsResponse wrapper for the ListMountTargets operation
type ListMountTargetsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []MountTargetSummary instances
	Items []MountTargetSummary `presentIn:"body"`

	// For pagination of a list of items. When paging through
	// a list, if this header appears in the response, then a
	// partial list might have been returned. Include this
	// value as the `page` parameter for the subsequent GET
	// request to get the next batch of items.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If
	// you need to contact Oracle about a particular request,
	// please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListMountTargetsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListMountTargetsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListMountTargetsLifecycleStateEnum Enum with underlying type: string
type ListMountTargetsLifecycleStateEnum string

// Set of constants representing the allowable values for ListMountTargetsLifecycleState
const (
	ListMountTargetsLifecycleStateCreating ListMountTargetsLifecycleStateEnum = "CREATING"
	ListMountTargetsLifecycleStateActive   ListMountTargetsLifecycleStateEnum = "ACTIVE"
	ListMountTargetsLifecycleStateDeleting ListMountTargetsLifecycleStateEnum = "DELETING"
	ListMountTargetsLifecycleStateDeleted  ListMountTargetsLifecycleStateEnum = "DELETED"
	ListMountTargetsLifecycleStateFailed   ListMountTargetsLifecycleStateEnum = "FAILED"
)

var mappingListMountTargetsLifecycleState = map[string]ListMountTargetsLifecycleStateEnum{
	"CREATING": ListMountTargetsLifecycleStateCreating,
	"ACTIVE":   ListMountTargetsLifecycleStateActive,
	"DELETING": ListMountTargetsLifecycleStateDeleting,
	"DELETED":  ListMountTargetsLifecycleStateDeleted,
	"FAILED":   ListMountTargetsLifecycleStateFailed,
}

// GetListMountTargetsLifecycleStateEnumValues Enumerates the set of values for ListMountTargetsLifecycleState
func GetListMountTargetsLifecycleStateEnumValues() []ListMountTargetsLifecycleStateEnum {
	values := make([]ListMountTargetsLifecycleStateEnum, 0)
	for _, v := range mappingListMountTargetsLifecycleState {
		values = append(values, v)
	}
	return values
}

// ListMountTargetsSortByEnum Enum with underlying type: string
type ListMountTargetsSortByEnum string

// Set of constants representing the allowable values for ListMountTargetsSortBy
const (
	ListMountTargetsSortByTimecreated ListMountTargetsSortByEnum = "TIMECREATED"
	ListMountTargetsSortByDisplayname ListMountTargetsSortByEnum = "DISPLAYNAME"
)

var mappingListMountTargetsSortBy = map[string]ListMountTargetsSortByEnum{
	"TIMECREATED": ListMountTargetsSortByTimecreated,
	"DISPLAYNAME": ListMountTargetsSortByDisplayname,
}

// GetListMountTargetsSortByEnumValues Enumerates the set of values for ListMountTargetsSortBy
func GetListMountTargetsSortByEnumValues() []ListMountTargetsSortByEnum {
	values := make([]ListMountTargetsSortByEnum, 0)
	for _, v := range mappingListMountTargetsSortBy {
		values = append(values, v)
	}
	return values
}

// ListMountTargetsSortOrderEnum Enum with underlying type: string
type ListMountTargetsSortOrderEnum string

// Set of constants representing the allowable values for ListMountTargetsSortOrder
const (
	ListMountTargetsSortOrderAsc  ListMountTargetsSortOrderEnum = "ASC"
	ListMountTargetsSortOrderDesc ListMountTargetsSortOrderEnum = "DESC"
)

var mappingListMountTargetsSortOrder = map[string]ListMountTargetsSortOrderEnum{
	"ASC":  ListMountTargetsSortOrderAsc,
	"DESC": ListMountTargetsSortOrderDesc,
}

// GetListMountTargetsSortOrderEnumValues Enumerates the set of values for ListMountTargetsSortOrder
func GetListMountTargetsSortOrderEnumValues() []ListMountTargetsSortOrderEnum {
	values := make([]ListMountTargetsSortOrderEnum, 0)
	for _, v := range mappingListMountTargetsSortOrder {
		values = append(values, v)
	}
	return values
}
