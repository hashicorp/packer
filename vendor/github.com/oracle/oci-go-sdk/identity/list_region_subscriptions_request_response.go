// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package identity

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListRegionSubscriptionsRequest wrapper for the ListRegionSubscriptions operation
type ListRegionSubscriptionsRequest struct {

	// The OCID of the tenancy.
	TenancyId *string `mandatory:"true" contributesTo:"path" name:"tenancyId"`
}

func (request ListRegionSubscriptionsRequest) String() string {
	return common.PointerString(request)
}

// ListRegionSubscriptionsResponse wrapper for the ListRegionSubscriptions operation
type ListRegionSubscriptionsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []RegionSubscription instance
	Items []RegionSubscription `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ListRegionSubscriptionsResponse) String() string {
	return common.PointerString(response)
}
