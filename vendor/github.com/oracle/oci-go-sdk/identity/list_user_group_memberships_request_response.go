// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

package identity

import (
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

// ListUserGroupMembershipsRequest wrapper for the ListUserGroupMemberships operation
type ListUserGroupMembershipsRequest struct {

	// The OCID of the compartment (remember that the tenancy is simply the root compartment).
	CompartmentId *string `mandatory:"true" contributesTo:"query" name:"compartmentId"`

	// The OCID of the user.
	UserId *string `mandatory:"false" contributesTo:"query" name:"userId"`

	// The OCID of the group.
	GroupId *string `mandatory:"false" contributesTo:"query" name:"groupId"`

	// The value of the `opc-next-page` response header from the previous "List" call.
	Page *string `mandatory:"false" contributesTo:"query" name:"page"`

	// The maximum number of items to return in a paginated "List" call.
	Limit *int `mandatory:"false" contributesTo:"query" name:"limit"`
}

func (request ListUserGroupMembershipsRequest) String() string {
	return common.PointerString(request)
}

// ListUserGroupMembershipsResponse wrapper for the ListUserGroupMemberships operation
type ListUserGroupMembershipsResponse struct {

	// The underlying http response
	RawResponse *http.Response

	// The []UserGroupMembership instance
	Items []UserGroupMembership `presentIn:"body"`

	// Unique Oracle-assigned identifier for the request. If you need to contact Oracle about a
	// particular request, please provide the request ID.
	OpcRequestId *string `presentIn:"header" name:"opc-request-id"`

	// For pagination of a list of items. When paging through a list, if this header appears in the response,
	// then a partial list might have been returned. Include this value as the `page` parameter for the
	// subsequent GET request to get the next batch of items.
	OpcNextPage *string `presentIn:"header" name:"opc-next-page"`
}

func (response ListUserGroupMembershipsResponse) String() string {
	return common.PointerString(response)
}
