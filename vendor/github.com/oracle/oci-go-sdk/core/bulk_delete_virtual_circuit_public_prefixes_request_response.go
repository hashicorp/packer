// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// BulkDeleteVirtualCircuitPublicPrefixesRequest wrapper for the BulkDeleteVirtualCircuitPublicPrefixes operation
type BulkDeleteVirtualCircuitPublicPrefixesRequest struct {

	// The OCID of the virtual circuit.
	VirtualCircuitId *string `mandatory:"true" contributesTo:"path" name:"virtualCircuitId"`

	// Request with publix prefixes to be deleted from the virtual circuit
	BulkDeleteVirtualCircuitPublicPrefixesDetails `contributesTo:"body"`
}

func (request BulkDeleteVirtualCircuitPublicPrefixesRequest) String() string {
	return common.PointerString(request)
}

// BulkDeleteVirtualCircuitPublicPrefixesResponse wrapper for the BulkDeleteVirtualCircuitPublicPrefixes operation
type BulkDeleteVirtualCircuitPublicPrefixesResponse struct {

	// The underlying http response
	RawResponse *http.Response
}

func (response BulkDeleteVirtualCircuitPublicPrefixesResponse) String() string {
	return common.PointerString(response)
}
