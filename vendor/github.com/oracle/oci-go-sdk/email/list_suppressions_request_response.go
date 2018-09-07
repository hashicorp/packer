// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package email

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListSuppressionsRequest wrapper for the ListSuppressions operation
type ListSuppressionsRequest struct {

	// The OCID for the compartment.
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The email address of the suppression.
	EmailAddress *string `mandatory:"false" contributesTo:"query" name:"emailAddress"`

	// Search for suppressions that were created within a specific date range,
	// using this parameter to specify the earliest creation date for the
	// returned list (inclusive). Specifying this parameter without the
	// corresponding `timeCreatedLessThan` parameter will retrieve suppressions created from the
	// given `timeCreatedGreaterThanOrEqualTo` to the current time, in "YYYY-MM-ddThh:mmZ" format with a
	// Z offset, as defined by RFC 3339.
	// **Example:** 2016-12-19T16:39:57.600Z
	TimeCreatedGreaterThanOrEqualTo *common.SDKTime `mandatory:"false" contributesTo:"query" name:"timeCreatedGreaterThanOrEqualTo"`

	// Search for suppressions that were created within a specific date range,
	// using this parameter to specify the latest creation date for the returned
	// list (exclusive). Specifying this parameter without the corresponding
	// `timeCreatedGreaterThanOrEqualTo` parameter will retrieve all suppressions created before the
	// specified end date, in "YYYY-MM-ddThh:mmZ" format with a Z offset, as
	// defined by RFC 3339.
	// **Example:** 2016-12-19T16:39:57.600Z
	TimeCreatedLessThan *common.SDKTime `mandatory:"false" contributesTo:"query" name:"timeCreatedLessThan"`

	// The value of the `opc-next-page` response header from the previous
	// GET request.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The maximum number of items to return in a paginated GET request.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`

	// The field to sort by. The `TIMECREATED` value returns the list in in
	// descending order by default. The `EMAILADDRESS` value returns the list in
	// ascending order by default. Use the `SortOrderQueryParam` to change the
	// direction of the returned list of items.
	SortBy ListSuppressionsSortByEnum `mandatory:"false" contributesTo:"query" name:"sortBy" omitEmpty:"true"`

	// The sort order to use, either ascending or descending order.
	SortOrder ListSuppressionsSortOrderEnum `mandatory:"false" contributesTo:"query" name:"sortOrder" omitEmpty:"true"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request ListSuppressionsRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request ListSuppressionsRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request ListSuppressionsRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// ListSuppressionsResponse wrapper for the ListSuppressions operation
type ListSuppressionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// A list of []SuppressionSummary instances
	Items []SuppressionSummary `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need
	// to contact Oracle about a particular request, please provide the
	// request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For pagination of a list of items. If this header appears in the
	// response, then a partial list might have been returned. Include
	// this value for the `page` parameter in subsequent GET
	// requests to return the next batch of items.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListSuppressionsResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response ListSuppressionsResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}

// ListSuppressionsSortByEnum Enum with underlying type: string
type ListSuppressionsSortByEnum string

// Set of constants representing the allowable values for ListSuppressionsSortBy
const (
	ListSuppressionsSortByTimecreated  ListSuppressionsSortByEnum = "TIMECREATED"
	ListSuppressionsSortByEmailaddress ListSuppressionsSortByEnum = "EMAILADDRESS"
)

var mappingListSuppressionsSortBy = map[string]ListSuppressionsSortByEnum{
	"TIMECREATED":  ListSuppressionsSortByTimecreated,
	"EMAILADDRESS": ListSuppressionsSortByEmailaddress,
}

// GetListSuppressionsSortByEnumValues Enumerates the set of values for ListSuppressionsSortBy
func GetListSuppressionsSortByEnumValues() []ListSuppressionsSortByEnum {
	values := make([]ListSuppressionsSortByEnum, 0)
	for _, v := range mappingListSuppressionsSortBy {
		values = append(values, v)
	}
	return values
}

// ListSuppressionsSortOrderEnum Enum with underlying type: string
type ListSuppressionsSortOrderEnum string

// Set of constants representing the allowable values for ListSuppressionsSortOrder
const (
	ListSuppressionsSortOrderAsc  ListSuppressionsSortOrderEnum = "ASC"
	ListSuppressionsSortOrderDesc ListSuppressionsSortOrderEnum = "DESC"
)

var mappingListSuppressionsSortOrder = map[string]ListSuppressionsSortOrderEnum{
	"ASC":  ListSuppressionsSortOrderAsc,
	"DESC": ListSuppressionsSortOrderDesc,
}

// GetListSuppressionsSortOrderEnumValues Enumerates the set of values for ListSuppressionsSortOrder
func GetListSuppressionsSortOrderEnumValues() []ListSuppressionsSortOrderEnum {
	values := make([]ListSuppressionsSortOrderEnum, 0)
	for _, v := range mappingListSuppressionsSortOrder {
		values = append(values, v)
	}
	return values
}
