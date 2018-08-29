// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package containerengine

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListClustersRequest wrapper for the ListClusters operation
type ListClustersRequest struct {

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// A cluster lifecycle state to filter on. Can have multiple parameters of this name.
	LifecycleState []ListClustersLifecycleStateEnum `contributesTo:"query" name:"lifecycleState" omitEmpty:"true" collectionFormat:"multi"`

	// The name to filter on.
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

	// The maximum number of items to return in a paginated "List" call.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The optional order in which to sort the results.
	SortOrder ListClustersSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the results by.
	SortBy ListClustersSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListClustersRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListClustersRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListClustersRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListClustersResponse wrapper for the ListClusters operation
type ListClustersResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ClusterSummary instances
	Items []ClusterSummary `presentIn:"body"`

	// For pagination of a list of items. When paging through a list, if this header appears in the response,
	// then there might be additional items still to get. Include this value as the `page` parameter for the
	// subsequent GET request.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListClustersResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListClustersResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListClustersLifecycleStateEnum Enum with underlying type: string
type ListClustersLifecycleStateEnum string

// Set of constants representing the allowable values for ListClustersLifecycleState
const (
	ListClustersLifecycleStateCreating ListClustersLifecycleStateEnum = "CREATING"
	ListClustersLifecycleStateActive   ListClustersLifecycleStateEnum = "ACTIVE"
	ListClustersLifecycleStateFailed   ListClustersLifecycleStateEnum = "FAILED"
	ListClustersLifecycleStateDeleting ListClustersLifecycleStateEnum = "DELETING"
	ListClustersLifecycleStateDeleted  ListClustersLifecycleStateEnum = "DELETED"
	ListClustersLifecycleStateUpdating ListClustersLifecycleStateEnum = "UPDATING"
)

var mappingListClustersLifecycleState = map[string]ListClustersLifecycleStateEnum{
	"CREATING": ListClustersLifecycleStateCreating,
	"ACTIVE":   ListClustersLifecycleStateActive,
	"FAILED":   ListClustersLifecycleStateFailed,
	"DELETING": ListClustersLifecycleStateDeleting,
	"DELETED":  ListClustersLifecycleStateDeleted,
	"UPDATING": ListClustersLifecycleStateUpdating,
}

// GetListClustersLifecycleStateEnumValues Enumerates the set of values for ListClustersLifecycleState
func GetListClustersLifecycleStateEnumValues() []ListClustersLifecycleStateEnum {
	values := make([]ListClustersLifecycleStateEnum, 0)
	for _, v := range mappingListClustersLifecycleState {
		values = append(values, v)
	}
	return values
}

// ListClustersSortOrderEnum Enum with underlying type: string
type ListClustersSortOrderEnum string

// Set of constants representing the allowable values for ListClustersSortOrder
const (
	ListClustersSortOrderAsc  ListClustersSortOrderEnum = "ASC"
	ListClustersSortOrderDesc ListClustersSortOrderEnum = "DESC"
)

var mappingListClustersSortOrder = map[string]ListClustersSortOrderEnum{
	"ASC":  ListClustersSortOrderAsc,
	"DESC": ListClustersSortOrderDesc,
}

// GetListClustersSortOrderEnumValues Enumerates the set of values for ListClustersSortOrder
func GetListClustersSortOrderEnumValues() []ListClustersSortOrderEnum {
	values := make([]ListClustersSortOrderEnum, 0)
	for _, v := range mappingListClustersSortOrder {
		values = append(values, v)
	}
	return values
}

// ListClustersSortByEnum Enum with underlying type: string
type ListClustersSortByEnum string

// Set of constants representing the allowable values for ListClustersSortBy
const (
	ListClustersSortById          ListClustersSortByEnum = "ID"
	ListClustersSortByName        ListClustersSortByEnum = "NAME"
	ListClustersSortByTimeCreated ListClustersSortByEnum = "TIME_CREATED"
)

var mappingListClustersSortBy = map[string]ListClustersSortByEnum{
	"ID":           ListClustersSortById,
	"NAME":         ListClustersSortByName,
	"TIME_CREATED": ListClustersSortByTimeCreated,
}

// GetListClustersSortByEnumValues Enumerates the set of values for ListClustersSortBy
func GetListClustersSortByEnumValues() []ListClustersSortByEnum {
	values := make([]ListClustersSortByEnum, 0)
	for _, v := range mappingListClustersSortBy {
		values = append(values, v)
	}
	return values
}
