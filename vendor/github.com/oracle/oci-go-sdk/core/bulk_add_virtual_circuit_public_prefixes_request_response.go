// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// BulkAddVirtualCircuitPublicPrefixesRequest wrapper for the BulkAddVirtualCircuitPublicPrefixes operation
type BulkAddVirtualCircuitPublicPrefixesRequest struct {

	// The OCID of the virtual circuit.
	VirtualCircuitId *string `mandatory:"true" contributesTo:"path" name:"virtualCircuitId"`

	// Request with publix prefixes to be added to the virtual circuit
	BulkAddVirtualCircuitPublicPrefixesDetails `contributesTo:"body"`
}

func (request BulkAddVirtualCircuitPublicPrefixesRequest) String() string {
	return common.PointerString(request)
}

// BulkAddVirtualCircuitPublicPrefixesResponse wrapper for the BulkAddVirtualCircuitPublicPrefixes operation
type BulkAddVirtualCircuitPublicPrefixesResponse struct {

	// The underlying http response
	RawResponse *http.Response
}

func (response BulkAddVirtualCircuitPublicPrefixesResponse) String() string {
	return common.PointerString(response)
}
