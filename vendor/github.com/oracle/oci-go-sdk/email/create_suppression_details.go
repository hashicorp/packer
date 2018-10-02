// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Email Delivery Service API
//
// API spec for managing OCI Email Delivery services.
//

package email

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateSuppressionDetails The details needed for creating a single suppression.
type CreateSuppressionDetails struct {

	// The OCID of the compartment to contain the suppression. Since
	// suppressions are at the customer level, this must be the tenancy
	// OCID.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// The recipient email address of the suppression.
	EmailAddress *string `mandatory:"false" json:"emailAddress"`
}

func (m CreateSuppressionDetails) String() string {
	return common.PointerString(m)
}
