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

// Vlan A resource to be used only with the Oracle Cloud VMware Solution.
// Conceptually, a virtual LAN (VLAN) is a broadcast domain that is created
// by partitioning and isolating a network at the data link layer (a *layer 2 network*).
// VLANs work by using IEEE 802.1Q VLAN tags. Layer 2 traffic is forwarded within the
// VLAN based on MAC learning.
// In the Networking service, a VLAN is an object within a VCN. You use VLANs to
// partition the VCN at the data link layer (layer 2). A VLAN is analagous to a subnet,
// which is an object for partitioning the VCN at the IP layer (layer 3).
type Vlan struct {

	// The range of IPv4 addresses that will be used for layer 3 communication with
	// hosts outside the VLAN.
	// Example: `192.168.1.0/24`
	CidrBlock *string `mandatory:"true" json:"cidrBlock"`

	// The OCID of the compartment containing the VLAN.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The VLAN's Oracle ID (OCID).
	Id *string `mandatory:"true" json:"id"`

	// The VLAN's current state.
	LifecycleState VlanLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID of the VCN the VLAN is in.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// The availability domain of the VLAN.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"false" json:"availabilityDomain"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// A list of the OCIDs of the network security groups (NSGs) to use with this VLAN.
	// All VNICs in the VLAN belong to these NSGs. For more
	// information about NSGs, see
	// NetworkSecurityGroup.
	NsgIds []string `mandatory:"false" json:"nsgIds"`

	// The IEEE 802.1Q VLAN tag of this VLAN.
	// Example: `100`
	VlanTag *int `mandatory:"false" json:"vlanTag"`

	// The OCID of the route table that the VLAN uses.
	RouteTableId *string `mandatory:"false" json:"routeTableId"`

	// The date and time the VLAN was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`
}

func (m Vlan) String() string {
	return common.PointerString(m)
}

// VlanLifecycleStateEnum Enum with underlying type: string
type VlanLifecycleStateEnum string

// Set of constants representing the allowable values for VlanLifecycleStateEnum
const (
	VlanLifecycleStateProvisioning VlanLifecycleStateEnum = "PROVISIONING"
	VlanLifecycleStateAvailable    VlanLifecycleStateEnum = "AVAILABLE"
	VlanLifecycleStateTerminating  VlanLifecycleStateEnum = "TERMINATING"
	VlanLifecycleStateTerminated   VlanLifecycleStateEnum = "TERMINATED"
	VlanLifecycleStateUpdating     VlanLifecycleStateEnum = "UPDATING"
)

var mappingVlanLifecycleState = map[string]VlanLifecycleStateEnum{
	"PROVISIONING": VlanLifecycleStateProvisioning,
	"AVAILABLE":    VlanLifecycleStateAvailable,
	"TERMINATING":  VlanLifecycleStateTerminating,
	"TERMINATED":   VlanLifecycleStateTerminated,
	"UPDATING":     VlanLifecycleStateUpdating,
}

// GetVlanLifecycleStateEnumValues Enumerates the set of values for VlanLifecycleStateEnum
func GetVlanLifecycleStateEnumValues() []VlanLifecycleStateEnum {
	values := make([]VlanLifecycleStateEnum, 0)
	for _, v := range mappingVlanLifecycleState {
		values = append(values, v)
	}
	return values
}
