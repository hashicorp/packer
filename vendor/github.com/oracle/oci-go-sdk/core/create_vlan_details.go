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

// CreateVlanDetails The representation of CreateVlanDetails
type CreateVlanDetails struct {

	// The availability domain of the VLAN.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The range of IPv4 addresses that will be used for layer 3 communication with
	// hosts outside the VLAN. The CIDR must maintain the following rules -
	// a. The CIDR block is valid and correctly formatted.
	// Example: `192.0.2.0/24`
	CidrBlock *string `mandatory:"true" json:"cidrBlock"`

	// The OCID of the compartment to contain the VLAN.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the VCN to contain the VLAN.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A descriptive name. Does not have to be unique, and it's changeable. Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// A list of the OCIDs of the network security groups (NSGs) to add all VNICs in the VLAN to. For more
	// information about NSGs, see
	// NetworkSecurityGroup.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// The OCID of the route table the VLAN will use. If you don't provide a value,
	// the VLAN uses the VCN's default route table.
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// The IEEE 802.1Q VLAN tag for this VLAN. The value must be unique across all
	// VLANs in the VCN. If you don't provide a value, Oracle assigns one.
	// You cannot change the value later. VLAN tag 0 is reserved for use by Oracle.
	VlanTag *int `mandatory:"false" json:"vlanTag"`
}

func (m CreateVlanDetails) String() string {
	return common.PointerString(m)
}
