// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package core

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// DeletePrivateIpRequest wrapper for the DeletePrivateIp operation
type DeletePrivateIpRequest struct {

	// The private IP's OCID.
	PrivateIpId *string `mandatory:"true" contributesTo:"path" name:"privateIpId"`

	// For optimistic concurrency control. In the PUT or DELETE call for a resource, set the `if-match`
	// parameter to the value of the etag from a previous GET or POST response for that resource.  The resource
	// will be updated or deleted only if the etag you provide matches the resource's current etag value.
	IfMatch *string `mandatory:"false" contributesTo:"header" name:"if-match"`
}

func (request DeletePrivateIpRequest) String() string {
	return common.PointerString(request)
}

// DeletePrivateIpResponse wrapper for the DeletePrivateIp operation
type DeletePrivateIpResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about
	// a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`
}

func (response DeletePrivateIpResponse) String() string {
	return common.PointerString(response)
}
