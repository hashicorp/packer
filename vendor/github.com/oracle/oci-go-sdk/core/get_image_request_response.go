// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetImageRequest wrapper for the GetImage operation
type GetImageRequest struct {

	// The OCID of the image.
	ImageId *string `mandatory:"true" contributesTo:"path" name:"imageId"`
}

func (request GetImageRequest) String() string {
	return common.PointerString(request)
}

// GetImageResponse wrapper for the GetImage operation
type GetImageResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Image instance
	Image `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetImageResponse) String() string {
	return common.PointerString(response)
}
