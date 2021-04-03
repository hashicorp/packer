// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
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
	"github.com/oracle/oci-go-sdk/v36/common"
)

// UpdateVlanDetails The representation of UpdateVlanDetails
type UpdateVlanDetails struct {

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A descriptive name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// A list of the OCIDs of the network security groups (NSGs) to use with
	// this VLAN. All VNICs in the VLAN will belong to these NSGs. For more
	// information about NSGs, see
	// NetworkSecurityGroup.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// The OCID of the route table the VLAN will use.
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// The CIDR block of the VLAN. The new CIDR block must meet the following criteria:
	// - Must be valid.
	// - The CIDR block's IP range must be completely within one of the VCN's CIDR block ranges.
	// - The old and new CIDR block ranges must use the same network address. Example: `10.0.0.0/25` and `10.0.0.0/24`.
	// - Must contain all IP addresses in use in the old CIDR range.
	// - The new CIDR range's broadcast address (last IP address of CIDR range) must not be an IP address in use in the old CIDR range.
	// **Note:** If you are changing the CIDR block, you cannot create VNICs or private IPs for this resource while the update is in progress.
	CidrBlock *string `mandatory:"false" json:"cidrBlock"`
}

func (m UpdateVlanDetails) String() string {
	return common.PointerString(m)
}
