// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetFastConnectProviderServiceRequest wrapper for the GetFastConnectProviderService operation
type GetFastConnectProviderServiceRequest struct {

	// The OCID of the provider service.
	ProviderServiceId *string `mandatory:"true" contributesTo:"path" name:"providerServiceId"`
}

func (request GetFastConnectProviderServiceRequest) String() string {
	return common.PointerString(request)
}

// GetFastConnectProviderServiceResponse wrapper for the GetFastConnectProviderService operation
type GetFastConnectProviderServiceResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The FastConnectProviderService instance
	FastConnectProviderService `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetFastConnectProviderServiceResponse) String() string {
	return common.PointerString(response)
}
