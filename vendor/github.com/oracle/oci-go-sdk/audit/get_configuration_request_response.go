// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package audit

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// GetConfigurationRequest wrapper for the GetConfiguration operation
type GetConfigurationRequest struct {

	// ID of the root compartment (tenancy)
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`
}

func (request GetConfigurationRequest) String() string {
	return common.PointerString(request)
}

// GetConfigurationResponse wrapper for the GetConfiguration operation
type GetConfigurationResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Configuration instance
	Configuration `presentIn:"body"`
}

func (response GetConfigurationResponse) String() string {
	return common.PointerString(response)
}
