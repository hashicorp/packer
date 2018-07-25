// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package identity

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// CreateRegionSubscriptionRequest wrapper for the CreateRegionSubscription operation
type CreateRegionSubscriptionRequest struct {

	// Request object for activate a new region.
	CreateRegionSubscriptionDetails `contributesTo:"body"`

	// The OCID of the tenancy.
	TenancyId *string `mandatory:"true" contributesTo:"path" name:"tenancyId"`

	// A token that uniquely identifies a request so it can be retried in case of a timeout or
	// server error without risk of executing that same action again. Retry tokens expire after 24
	// hours, but can be invalidated before then due to conflicting operations (e.g., if a resource
	// has been deleted and purged from the system, then a retry of the original creation request
	// may be rejected).
	OpcRetryToken *string `mandatory:"false" contributesTo:"header" name:"opc-retry-token"`
}

func (request CreateRegionSubscriptionRequest) String() string {
	return common.PointerString(request)
}

// CreateRegionSubscriptionResponse wrapper for the CreateRegionSubscription operation
type CreateRegionSubscriptionResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The RegionSubscription instance
	RegionSubscription `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response CreateRegionSubscriptionResponse) String() string {
	return common.PointerString(response)
}
