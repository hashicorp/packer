// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package dns

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// UpdateZoneRequest wrapper for the UpdateZone operation
type UpdateZoneRequest struct {

	// The name or OCID of the target zone.
	ZoneNameOrId *string `mandatory:"true" contributesTo:"path" name:"zoneNameOrId"`

	// New data for the zone.
	UpdateZoneDetails `contributesTo:"body"`

	// The `If-Match` header field makes the request method conditional on the
	// existence of at least one current representation of the target resource,
	// when the field-value is `*`, or having a current representation of the
	// target resource that has an entity-tag matching a member of the list of
	// entity-tags provided in the field-value.
	IfMatch *string `mandatory:"false" contributesTo:"header" name:"If-Match"`

	// The `If-Unmodified-Since` header field makes the request method
	// conditional on the selected representation's last modification date being
	// earlier than or equal to the date provided in the field-value.  This
	// field accomplishes the same purpose as If-Match for cases where the user
	// agent does not have an entity-tag for the representation.
	IfUnmodifiedSince *string `mandatory:"false" contributesTo:"header" name:"If-Unmodified-Since"`

	// The OCID of the compartment the resource belongs to.
	CompartmentId *string `mandatory:"false" contributesTo:"query" name:"compartmentId"`

	// Unique Oracle-assigned identifier for the request.
	// If you need to contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" contributesTo:"header" name:"opc-request-id"`

	// Metadata about the request. This information will not be transmitted to the service, but
	// represents information that the SDK will consume to drive retry behavior.
	RequestMetadata common.RequestMetadata
}

func (request UpdateZoneRequest) String() string {
	return common.PointerString(request)
}

// HTTPRequest implements the OCIRequest interface
func (request UpdateZoneRequest) HTTPRequest(method, path string) (http.Request, error) {
	return common.MakeDefaultHTTPRequestWithTaggedStruct(method, path, request)
}

// RetryPolicy implements the OCIRetryableRequest interface. This retrieves the specified retry policy.
func (request UpdateZoneRequest) RetryPolicy() *common.RetryPolicy {
	return request.RequestMetadata.RetryPolicy
}

// UpdateZoneResponse wrapper for the UpdateZone operation
type UpdateZoneResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The Zone instance
	Zone `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to
	// contact Oracle about a particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// The current version of the zone, ending with a
	// representation-specific suffix. This value may be used in If-Match
	// and If-None-Match headers for later requests of the same resource.
	ETag *string `presentIn:"header" name:"etag"`
}

func (response UpdateZoneResponse) String() string {
	return common.PointerString(response)
}

// HTTPResponse implements the OCIResponse interface
func (response UpdateZoneResponse) HTTPResponse() *http.Response {
	return response.RawResponse
}
