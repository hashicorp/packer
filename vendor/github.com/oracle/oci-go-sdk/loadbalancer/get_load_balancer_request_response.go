// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package loadbalancer

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetLoadBalancerRequest wrapper for the GetLoadBalancer operation
type GetLoadBalancerRequest struct {

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm) of the load balancer to retrieve.
	LoadBalancerId *string `mandatory:"true" contributesTo:"path" name:"loadBalancerId"`

	// The unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`
}

func (request GetLoadBalancerRequest) String() string {
	return common.PointerString(request)
}

// GetLoadBalancerResponse wrapper for the GetLoadBalancer operation
type GetLoadBalancerResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The LoadBalancer instance
	LoadBalancer `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetLoadBalancerResponse) String() string {
	return common.PointerString(response)
}
