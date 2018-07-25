// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetVolumeRequest wrapper for the GetVolume operation
type GetVolumeRequest struct {

	// The OCID of the volume.
	VolumeId *string `mandatory:"true" contributesTo:"path" name:"volumeId"`
}

func (request GetVolumeRequest) String() string {
	return common.PointerString(request)
}

// GetVolumeResponse wrapper for the GetVolume operation
type GetVolumeResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Volume instance
	Volume `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetVolumeResponse) String() string {
	return common.PointerString(response)
}
