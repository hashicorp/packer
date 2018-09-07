// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package dns

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListZonesRequest wrapper for the ListZones operation
type ListZonesRequest struct {

	// The OCID of the compartment the resource belongs to.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The maximum number of items to return in a page of the collection.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// A case-sensitive filter for zone names.
	// Will match any zone with a name that equals the provided value.
	Name *string `mandatory:"false" contributesTo:"query" name:"name"`

	// Search by zone name.
	// Will match any zone whose name (case-insensitive) contains the provided value.
	NameContains *string `mandatory:"false" contributesTo:"query" name:"nameContains"`

	// Search by zone type, `PRIMARY` or `SECONDARY`.
	// Will match any zone whose type equals the provided value.
	ZoneType ListZonesZoneTypeEnum `mandatory:"false" contributesTo:"query" name:"zoneType" omitEmpty:"true"`

	// An RFC 3339 (https://www.ietf.org/rfc/rfc3339.txt) timestamp that states
	// all returned resources were created on or after the indicated time.
	TimeCreatedGreaterThanOrEqualTo *common.SDKTime `mandatory:"false" contributesTo:"query" name:"timeCreatedGreaterThanOrEqualTo"`

	// An RFC 3339 (https://www.ietf.org/rfc/rfc3339.txt) timestamp that states
	// all returned resources were created before the indicated time.
	TimeCreatedLessThan *common.SDKTime `mandatory:"false" contributesTo:"query" name:"timeCreatedLessThan"`

	// The field by which to sort zones.
	SortBy ListZonesSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The order to sort the resources.
	SortOrder ListZonesSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// The state of a resource.
	LifecycleState ListZonesLifecycleStateEnum `mandatory:"false" contributesTo:"query" name:"lifecycleState" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListZonesRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListZonesRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListZonesRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListZonesResponse wrapper for the ListZones operation
type ListZonesResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []ZoneSummary instances
	Items []ZoneSummary `presentIn:"body"`

	// For pagination of a list of items. When paging through a list, if
	// this header appears in the response, then a partial list might have
	// been returned. Include this value as the `page` parameter for the
	// subsequent GET request to get the next batch of items.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`

	// The total number of items that match the query.
	OpcTotalItems *int `presentIn:"header" name:"opc-total-items"`

	// Unique Oracle-assigned identifier for the request. If you need to
	// contact Oracle about a particular request, please provide the request
	// ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListZonesResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListZonesResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListZonesZoneTypeEnum Enum with underlying type: string
type ListZonesZoneTypeEnum string

// Set of constants representing the allowable values for ListZonesZoneType
const (
	ListZonesZoneTypePrimary   ListZonesZoneTypeEnum = "PRIMARY"
	ListZonesZoneTypeSecondary ListZonesZoneTypeEnum = "SECONDARY"
)

var mappingListZonesZoneType = map[string]ListZonesZoneTypeEnum{
	"PRIMARY":   ListZonesZoneTypePrimary,
	"SECONDARY": ListZonesZoneTypeSecondary,
}

// GetListZonesZoneTypeEnumValues Enumerates the set of values for ListZonesZoneType
func GetListZonesZoneTypeEnumValues() []ListZonesZoneTypeEnum {
	values := make([]ListZonesZoneTypeEnum, 0)
	for _, v := range mappingListZonesZoneType {
		values = append(values, v)
	}
	return values
}

// ListZonesSortByEnum Enum with underlying type: string
type ListZonesSortByEnum string

// Set of constants representing the allowable values for ListZonesSortBy
const (
	ListZonesSortByName        ListZonesSortByEnum = "name"
	ListZonesSortByZonetype    ListZonesSortByEnum = "zoneType"
	ListZonesSortByTimecreated ListZonesSortByEnum = "timeCreated"
)

var mappingListZonesSortBy = map[string]ListZonesSortByEnum{
	"name":        ListZonesSortByName,
	"zoneType":    ListZonesSortByZonetype,
	"timeCreated": ListZonesSortByTimecreated,
}

// GetListZonesSortByEnumValues Enumerates the set of values for ListZonesSortBy
func GetListZonesSortByEnumValues() []ListZonesSortByEnum {
	values := make([]ListZonesSortByEnum, 0)
	for _, v := range mappingListZonesSortBy {
		values = append(values, v)
	}
	return values
}

// ListZonesSortOrderEnum Enum with underlying type: string
type ListZonesSortOrderEnum string

// Set of constants representing the allowable values for ListZonesSortOrder
const (
	ListZonesSortOrderAsc  ListZonesSortOrderEnum = "ASC"
	ListZonesSortOrderDesc ListZonesSortOrderEnum = "DESC"
)

var mappingListZonesSortOrder = map[string]ListZonesSortOrderEnum{
	"ASC":  ListZonesSortOrderAsc,
	"DESC": ListZonesSortOrderDesc,
}

// GetListZonesSortOrderEnumValues Enumerates the set of values for ListZonesSortOrder
func GetListZonesSortOrderEnumValues() []ListZonesSortOrderEnum {
	values := make([]ListZonesSortOrderEnum, 0)
	for _, v := range mappingListZonesSortOrder {
		values = append(values, v)
	}
	return values
}

// ListZonesLifecycleStateEnum Enum with underlying type: string
type ListZonesLifecycleStateEnum string

// Set of constants representing the allowable values for ListZonesLifecycleState
const (
	ListZonesLifecycleStateActive   ListZonesLifecycleStateEnum = "ACTIVE"
	ListZonesLifecycleStateCreating ListZonesLifecycleStateEnum = "CREATING"
	ListZonesLifecycleStateDeleted  ListZonesLifecycleStateEnum = "DELETED"
	ListZonesLifecycleStateDeleting ListZonesLifecycleStateEnum = "DELETING"
	ListZonesLifecycleStateFailed   ListZonesLifecycleStateEnum = "FAILED"
)

var mappingListZonesLifecycleState = map[string]ListZonesLifecycleStateEnum{
	"ACTIVE":   ListZonesLifecycleStateActive,
	"CREATING": ListZonesLifecycleStateCreating,
	"DELETED":  ListZonesLifecycleStateDeleted,
	"DELETING": ListZonesLifecycleStateDeleting,
	"FAILED":   ListZonesLifecycleStateFailed,
}

// GetListZonesLifecycleStateEnumValues Enumerates the set of values for ListZonesLifecycleState
func GetListZonesLifecycleStateEnumValues() []ListZonesLifecycleStateEnum {
	values := make([]ListZonesLifecycleStateEnum, 0)
	for _, v := range mappingListZonesLifecycleState {
		values = append(values, v)
	}
	return values
}
