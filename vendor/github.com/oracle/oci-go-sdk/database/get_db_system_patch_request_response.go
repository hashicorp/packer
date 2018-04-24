// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package database

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetDbSystemPatchRequest wrapper for the GetDbSystemPatch operation
type GetDbSystemPatchRequest struct {

	// The DB System OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm).
	DbSystemId *string `mandatory:"true" contributesTo:"path" name:"dbSystemId"`

	// The OCID of the patch.
	PatchId *string `mandatory:"true" contributesTo:"path" name:"patchId"`
}

func (request GetDbSystemPatchRequest) String() string {
	return common.PointerString(request)
}

// GetDbSystemPatchResponse wrapper for the GetDbSystemPatch operation
type GetDbSystemPatchResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Patch instance
	Patch `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetDbSystemPatchResponse) String() string {
	return common.PointerString(response)
}
