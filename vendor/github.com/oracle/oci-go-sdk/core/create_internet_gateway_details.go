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

// CreateInternetGatewayDetails The representation of CreateInternetGatewayDetails
type CreateInternetGatewayDetails struct {

	// The OCID of the compartment to contain the Internet Gateway.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Whether the gateway is enabled upon creation.
	IsEnabled *bool `mandatory:"true" json:"isEnabled"`

	// The OCID of the VCN the Internet Gateway is attached to.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// A user-friendly name. Does not have to be unique, and it's changeable. Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m CreateInternetGatewayDetails) String() string {
	return common.PointerString(m)
}
