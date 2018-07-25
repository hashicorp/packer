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

// CreateLocalPeeringGatewayDetails The representation of CreateLocalPeeringGatewayDetails
type CreateLocalPeeringGatewayDetails struct {

	// The OCID of the compartment containing the local peering gateway (LPG).
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the VCN the LPG belongs to.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// A user-friendly name. Does not have to be unique, and it's changeable. Avoid
	// entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m CreateLocalPeeringGatewayDetails) String() string {
	return common.PointerString(m)
}
