// Copyright (c) 2016, 2018, 2020, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
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
	// Example: `203.0.113.2`
	IpAddress *string `mandatory:"true" json:"ipAddress"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable. Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the CPE device type. You can provide
	// a value if you want to later generate CPE device configuration content for IPSec connections
	// that use this CPE. You can also call UpdateCpe later to
	// provide a value. For a list of possible values, see
	// ListCpeDeviceShapes.
	// For more information about generating CPE device configuration content, see:
	//   * GetCpeDeviceConfigContent
	//   * GetIpsecCpeDeviceConfigContent
	//   * GetTunnelCpeDeviceConfigContent
	//   * GetTunnelCpeDeviceConfig
	CpeDeviceShapeId *string `mandatory:"false" json:"cpeDeviceShapeId"`
}

func (m CreateCpeDetails) String() string {
	return common.PointerString(m)
}
