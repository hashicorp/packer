// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"github.com/oracle/oci-go-sdk/common"
)

// UpdateInternetGatewayDetails The representation of UpdateInternetGatewayDetails
type UpdateInternetGatewayDetails struct {

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Whether the gateway is enabled.
	IsEnabled *bool `mandatory:"false" json:"isEnabled"`
}

func (m UpdateInternetGatewayDetails) String() string {
	return common.PointerString(m)
}
