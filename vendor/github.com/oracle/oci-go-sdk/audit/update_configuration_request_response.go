// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package audit

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// UpdateConfigurationRequest wrapper for the UpdateConfiguration operation
type UpdateConfigurationRequest struct {

	// ID of the root compartment (tenancy)
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The configuration properties
	UpdateConfigurationDetails `contributesTo:"body"`
}

func (request UpdateConfigurationRequest) String() string {
	return common.PointerString(request)
}

// UpdateConfigurationResponse wrapper for the UpdateConfiguration operation
type UpdateConfigurationResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm) of the work request.
	OpcWorkRequestId *string `presentIn:"header" name:"opc-work-request-id"`
}

func (response UpdateConfigurationResponse) String() string {
	return common.PointerString(response)
}
