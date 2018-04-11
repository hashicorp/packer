// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetVolumeBackupRequest wrapper for the GetVolumeBackup operation
type GetVolumeBackupRequest struct {

	// The OCID of the volume backup.
	VolumeBackupId *string `mandatory:"true" contributesTo:"path" name:"volumeBackupId"`
}

func (request GetVolumeBackupRequest) String() string {
	return common.PointerString(request)
}

// GetVolumeBackupResponse wrapper for the GetVolumeBackup operation
type GetVolumeBackupResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The VolumeBackup instance
	VolumeBackup `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetVolumeBackupResponse) String() string {
	return common.PointerString(response)
}
