// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package identity

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListApiKeysRequest wrapper for the ListApiKeys operation
type ListApiKeysRequest struct {

	// The OCID of the user.
	UserId *string `mandatory:"true" contributesTo:"path" name:"userId"`
}

func (request ListApiKeysRequest) String() string {
	return common.PointerString(request)
}

// ListApiKeysResponse wrapper for the ListApiKeys operation
type ListApiKeysResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []ApiKey instance
	Items []ApiKey `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For pagination of a list of items. When paging through a list, if this header appears in the response,
	// then a partial list might have been returned. Include this value as the `page` parameter for the
	// subsequent GET request to get the next batch of items.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListApiKeysResponse) String() string {
	return common.PointerString(response)
}
