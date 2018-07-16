// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package containerengine

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListWorkRequestsRequest wrapper for the ListWorkRequests operation
type ListWorkRequestsRequest struct {

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID of the cluster.
	ClusterId *string `mandatory:"false" contributesTo:"query" name:"clusterId"`

	// The OCID of the resource associated with a work request
	ResourceId *string `mandatory:"false" contributesTo:"query" name:"resourceId"`

	// Type of the resource associated with a work request
	ResourceType ListWorkRequestsResourceTypeEnum `mandatory:"false" contributesTo:"query" name:"resourceType" omitEmpty:"true"`

	// A work request status to filter on. Can have multiple parameters of this name.
	Status []ListWorkRequestsStatusEnum `contributesTo:"query" name:"status" omitEmpty:"true" collectionFormat:"multi"`

	// The maximum number of items to return in a paginated "List" call.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The optional order in which to sort the results.
	SortOrder ListWorkRequestsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The optional field to sort the results by.
	SortBy ListWorkRequestsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request. If you need to contact
	// Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListWorkRequestsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListWorkRequestsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListWorkRequestsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListWorkRequestsResponse wrapper for the ListWorkRequests operation
type ListWorkRequestsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []WorkRequestSummary instances
	Items []WorkRequestSummary `presentIn:"body"`

	// For pagination of a list of items. When paging through a list, if this header appears in the response,
	// then there might be additional items still to get. Include this value as the `page` parameter for the
	// subsequent GET request.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListWorkRequestsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListWorkRequestsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListWorkRequestsResourceTypeEnum Enum with underlying type: string
type ListWorkRequestsResourceTypeEnum string

// Set of constants representing the allowable values for ListWorkRequestsResourceType
const (
	ListWorkRequestsResourceTypeCluster  ListWorkRequestsResourceTypeEnum = "CLUSTER"
	ListWorkRequestsResourceTypeNodepool ListWorkRequestsResourceTypeEnum = "NODEPOOL"
)

var mappingListWorkRequestsResourceType = map[string]ListWorkRequestsResourceTypeEnum{
	"CLUSTER":  ListWorkRequestsResourceTypeCluster,
	"NODEPOOL": ListWorkRequestsResourceTypeNodepool,
}

// GetListWorkRequestsResourceTypeEnumValues Enumerates the set of values for ListWorkRequestsResourceType
func GetListWorkRequestsResourceTypeEnumValues() []ListWorkRequestsResourceTypeEnum {
	values := make([]ListWorkRequestsResourceTypeEnum, 0)
	for _, v := range mappingListWorkRequestsResourceType {
		values = append(values, v)
	}
	return values
}

// ListWorkRequestsStatusEnum Enum with underlying type: string
type ListWorkRequestsStatusEnum string

// Set of constants representing the allowable values for ListWorkRequestsStatus
const (
	ListWorkRequestsStatusAccepted   ListWorkRequestsStatusEnum = "ACCEPTED"
	ListWorkRequestsStatusInProgress ListWorkRequestsStatusEnum = "IN_PROGRESS"
	ListWorkRequestsStatusFailed     ListWorkRequestsStatusEnum = "FAILED"
	ListWorkRequestsStatusSucceeded  ListWorkRequestsStatusEnum = "SUCCEEDED"
	ListWorkRequestsStatusCanceling  ListWorkRequestsStatusEnum = "CANCELING"
	ListWorkRequestsStatusCanceled   ListWorkRequestsStatusEnum = "CANCELED"
)

var mappingListWorkRequestsStatus = map[string]ListWorkRequestsStatusEnum{
	"ACCEPTED":    ListWorkRequestsStatusAccepted,
	"IN_PROGRESS": ListWorkRequestsStatusInProgress,
	"FAILED":      ListWorkRequestsStatusFailed,
	"SUCCEEDED":   ListWorkRequestsStatusSucceeded,
	"CANCELING":   ListWorkRequestsStatusCanceling,
	"CANCELED":    ListWorkRequestsStatusCanceled,
}

// GetListWorkRequestsStatusEnumValues Enumerates the set of values for ListWorkRequestsStatus
func GetListWorkRequestsStatusEnumValues() []ListWorkRequestsStatusEnum {
	values := make([]ListWorkRequestsStatusEnum, 0)
	for _, v := range mappingListWorkRequestsStatus {
		values = append(values, v)
	}
	return values
}

// ListWorkRequestsSortOrderEnum Enum with underlying type: string
type ListWorkRequestsSortOrderEnum string

// Set of constants representing the allowable values for ListWorkRequestsSortOrder
const (
	ListWorkRequestsSortOrderAsc  ListWorkRequestsSortOrderEnum = "ASC"
	ListWorkRequestsSortOrderDesc ListWorkRequestsSortOrderEnum = "DESC"
)

var mappingListWorkRequestsSortOrder = map[string]ListWorkRequestsSortOrderEnum{
	"ASC":  ListWorkRequestsSortOrderAsc,
	"DESC": ListWorkRequestsSortOrderDesc,
}

// GetListWorkRequestsSortOrderEnumValues Enumerates the set of values for ListWorkRequestsSortOrder
func GetListWorkRequestsSortOrderEnumValues() []ListWorkRequestsSortOrderEnum {
	values := make([]ListWorkRequestsSortOrderEnum, 0)
	for _, v := range mappingListWorkRequestsSortOrder {
		values = append(values, v)
	}
	return values
}

// ListWorkRequestsSortByEnum Enum with underlying type: string
type ListWorkRequestsSortByEnum string

// Set of constants representing the allowable values for ListWorkRequestsSortBy
const (
	ListWorkRequestsSortById            ListWorkRequestsSortByEnum = "ID"
	ListWorkRequestsSortByOperationType ListWorkRequestsSortByEnum = "OPERATION_TYPE"
	ListWorkRequestsSortByStatus        ListWorkRequestsSortByEnum = "STATUS"
	ListWorkRequestsSortByTimeAccepted  ListWorkRequestsSortByEnum = "TIME_ACCEPTED"
	ListWorkRequestsSortByTimeStarted   ListWorkRequestsSortByEnum = "TIME_STARTED"
	ListWorkRequestsSortByTimeFinished  ListWorkRequestsSortByEnum = "TIME_FINISHED"
)

var mappingListWorkRequestsSortBy = map[string]ListWorkRequestsSortByEnum{
	"ID":             ListWorkRequestsSortById,
	"OPERATION_TYPE": ListWorkRequestsSortByOperationType,
	"STATUS":         ListWorkRequestsSortByStatus,
	"TIME_ACCEPTED":  ListWorkRequestsSortByTimeAccepted,
	"TIME_STARTED":   ListWorkRequestsSortByTimeStarted,
	"TIME_FINISHED":  ListWorkRequestsSortByTimeFinished,
}

// GetListWorkRequestsSortByEnumValues Enumerates the set of values for ListWorkRequestsSortBy
func GetListWorkRequestsSortByEnumValues() []ListWorkRequestsSortByEnum {
	values := make([]ListWorkRequestsSortByEnum, 0)
	for _, v := range mappingListWorkRequestsSortBy {
		values = append(values, v)
	}
	return values
}
