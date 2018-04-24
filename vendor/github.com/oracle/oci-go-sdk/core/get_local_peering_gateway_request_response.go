// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetLocalPeeringGatewayRequest wrapper for the GetLocalPeeringGateway operation
type GetLocalPeeringGatewayRequest struct {

	// The OCID of the local peering gateway.
	LocalPeeringGatewayId *string `mandatory:"true" contributesTo:"path" name:"localPeeringGatewayId"`
}

func (request GetLocalPeeringGatewayRequest) String() string {
	return common.PointerString(request)
}

// GetLocalPeeringGatewayResponse wrapper for the GetLocalPeeringGateway operation
type GetLocalPeeringGatewayResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The LocalPeeringGateway instance
	LocalPeeringGateway `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetLocalPeeringGatewayResponse) String() string {
	return common.PointerString(response)
}
