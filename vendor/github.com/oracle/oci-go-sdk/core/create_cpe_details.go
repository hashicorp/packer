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

// CreateCpeDetails The representation of CreateCpeDetails
type CreateCpeDetails struct {

	// The OCID of the compartment to contain the CPE.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The public IP address of the on-premises router.
	// Example: `143.19.23.16`
	IpAddress *string `mandatory:"true" json:"ipAddress"`

	// A user-friendly name. Does not have to be unique, and it's changeable. Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m CreateCpeDetails) String() string {
	return common.PointerString(m)
}
