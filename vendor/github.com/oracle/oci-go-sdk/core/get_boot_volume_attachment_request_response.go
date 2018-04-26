// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetBootVolumeAttachmentRequest wrapper for the GetBootVolumeAttachment operation
type GetBootVolumeAttachmentRequest struct {

	// The OCID of the boot volume attachment.
	BootVolumeAttachmentId *string `mandatory:"true" contributesTo:"path" name:"bootVolumeAttachmentId"`
}

func (request GetBootVolumeAttachmentRequest) String() string {
	return common.PointerString(request)
}

// GetBootVolumeAttachmentResponse wrapper for the GetBootVolumeAttachment operation
type GetBootVolumeAttachmentResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The BootVolumeAttachment instance
	BootVolumeAttachment `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetBootVolumeAttachmentResponse) String() string {
	return common.PointerString(response)
}
