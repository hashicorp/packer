// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package database

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetBackupRequest wrapper for the GetBackup operation
type GetBackupRequest struct {

	// The backup OCID.
	BackupId *string `mandatory:"true" contributesTo:"path" name:"backupId"`
}

func (request GetBackupRequest) String() string {
	return common.PointerString(request)
}

// GetBackupResponse wrapper for the GetBackup operation
type GetBackupResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Backup instance
	Backup `presentIn:"body"`

	// For optimistic concurrency control. See `if-match`.
	Etag *string `presentIn:"header" name:"etag"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response GetBackupResponse) String() string {
	return common.PointerString(response)
}
