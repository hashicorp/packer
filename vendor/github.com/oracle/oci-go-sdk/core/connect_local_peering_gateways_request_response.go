// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ConnectLocalPeeringGatewaysRequest wrapper for the ConnectLocalPeeringGateways operation
type ConnectLocalPeeringGatewaysRequest struct {

	// The OCID of the local peering gateway.
	LocalPeeringGatewayId *string `mandatory:"true" contributesTo:"path" name:"localPeeringGatewayId"`

	// Details regarding the local peering gateway to connect.
	ConnectLocalPeeringGatewaysDetails `contributesTo:"body"`
}

func (request ConnectLocalPeeringGatewaysRequest) String() string {
	return common.PointerString(request)
}

// ConnectLocalPeeringGatewaysResponse wrapper for the ConnectLocalPeeringGateways operation
type ConnectLocalPeeringGatewaysResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response ConnectLocalPeeringGatewaysResponse) String() string {
	return common.PointerString(response)
}
