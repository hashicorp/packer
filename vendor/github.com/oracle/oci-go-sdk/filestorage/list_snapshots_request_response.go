// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package filestorage

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListSnapshotsRequest wrapper for the ListSnapshots operation
type ListSnapshotsRequest struct {

	// The OCID of the file system.
	FileSystemId *string `mandatory:"true" contributesTo:"query" name:"fileSystemId"`

	// The maximum number of items to return in a paginated "List" call.
	// Example: `500`
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// Filter results by the specified lifecycle state. Must be a valid
	// state for the resource type.
	LifecycleState ListSnapshotsLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Filter results by OCID. Must be an OCID of the correct type for
	// the resouce type.
	Id *string `mandatory:"false" contributesTo:"query" name:"id"`

	// The sort order to use, either 'asc' or 'desc', where 'asc' is
	// ascending and 'desc' is descending.
	SortOrder ListSnapshotsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListSnapshotsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListSnapshotsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListSnapshotsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListSnapshotsResponse wrapper for the ListSnapshots operation
type ListSnapshotsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []SnapshotSummary instances
	Items []SnapshotSummary `presentIn:"body"`

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

func (response ListSnapshotsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListSnapshotsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListSnapshotsLifecycleStateEnum Enum with underlying type: string
type ListSnapshotsLifecycleStateEnum string

// Set of constants representing the allowable values for ListSnapshotsLifecycleState
const (
	ListSnapshotsLifecycleStateCreating ListSnapshotsLifecycleStateEnum = "CREATING"
	ListSnapshotsLifecycleStateActive   ListSnapshotsLifecycleStateEnum = "ACTIVE"
	ListSnapshotsLifecycleStateDeleting ListSnapshotsLifecycleStateEnum = "DELETING"
	ListSnapshotsLifecycleStateDeleted  ListSnapshotsLifecycleStateEnum = "DELETED"
	ListSnapshotsLifecycleStateFailed   ListSnapshotsLifecycleStateEnum = "FAILED"
)

var mappingListSnapshotsLifecycleState = map[string]ListSnapshotsLifecycleStateEnum{
	"CREATING": ListSnapshotsLifecycleStateCreating,
	"ACTIVE":   ListSnapshotsLifecycleStateActive,
	"DELETING": ListSnapshotsLifecycleStateDeleting,
	"DELETED":  ListSnapshotsLifecycleStateDeleted,
	"FAILED":   ListSnapshotsLifecycleStateFailed,
}

// GetListSnapshotsLifecycleStateEnumValues Enumerates the set of values for ListSnapshotsLifecycleState
func GetListSnapshotsLifecycleStateEnumValues() []ListSnapshotsLifecycleStateEnum {
	values := make([]ListSnapshotsLifecycleStateEnum, 0)
	for _, v := range mappingListSnapshotsLifecycleState {
		values = append(values, v)
	}
	return values
}

// ListSnapshotsSortOrderEnum Enum with underlying type: string
type ListSnapshotsSortOrderEnum string

// Set of constants representing the allowable values for ListSnapshotsSortOrder
const (
	ListSnapshotsSortOrderAsc  ListSnapshotsSortOrderEnum = "ASC"
	ListSnapshotsSortOrderDesc ListSnapshotsSortOrderEnum = "DESC"
)

var mappingListSnapshotsSortOrder = map[string]ListSnapshotsSortOrderEnum{
	"ASC":  ListSnapshotsSortOrderAsc,
	"DESC": ListSnapshotsSortOrderDesc,
}

// GetListSnapshotsSortOrderEnumValues Enumerates the set of values for ListSnapshotsSortOrder
func GetListSnapshotsSortOrderEnumValues() []ListSnapshotsSortOrderEnum {
	values := make([]ListSnapshotsSortOrderEnum, 0)
	for _, v := range mappingListSnapshotsSortOrder {
		values = append(values, v)
	}
	return values
}
