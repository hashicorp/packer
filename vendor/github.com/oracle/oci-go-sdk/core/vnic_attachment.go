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

// VnicAttachment Represents an attachment between a VNIC and an instance. For more information, see
// Virtual Network Interface Cards (VNICs) (https://docs.cloud.oracle.com/Content/Network/Tasks/managingVNICs.htm).
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
type VnicAttachment struct {

	// The availability domain of the instance.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment the VNIC attachment is in, which is the same
	// compartment the instance is in.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the VNIC attachment.
	Id *string `mandatory:"true" json:"id"`

	// The OCID of the instance.
	InstanceId *string `mandatory:"true" json:"instanceId"`

	// The current state of the VNIC attachment.
	LifecycleState VnicAttachmentLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The date and time the VNIC attachment was created, in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// A user-friendly name. Does not have to be unique.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Which physical network interface card (NIC) the VNIC uses.
	// Certain bare metal instance shapes have two active physical NICs (0 and 1). If
	// you add a secondary VNIC to one of these instances, you can specify which NIC
	// the VNIC will use. For more information, see
	// Virtual Network Interface Cards (VNICs) (https://docs.cloud.oracle.com/Content/Network/Tasks/managingVNICs.htm).
	NicIndex *int `mandatory:"false" json:"nicIndex"`

	// The OCID of the subnet to create the VNIC in.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// The OCID of the VLAN to create the VNIC in. Creating the VNIC in a VLAN (instead
	// of a subnet) is possible only if you are an Oracle Cloud VMware Solution customer.
	// See Vlan.
	// An error is returned if the instance already has a VNIC attached to it from this VLAN.
	VlanId *string `mandatory:"false" json:"vlanId"`

	// The Oracle-assigned VLAN tag of the attached VNIC. Available after the
	// attachment process is complete.
	// However, if the VNIC belongs to a VLAN as part of the Oracle Cloud VMware Solution,
	// the `vlanTag` value is instead the value of the `vlanTag` attribute for the VLAN.
	// See Vlan.
	// Example: `0`
	VlanTag *int `mandatory:"false" json:"vlanTag"`

	// The OCID of the VNIC. Available after the attachment process is complete.
	VnicId *string `mandatory:"false" json:"vnicId"`
}

func (m VnicAttachment) String() string {
	return common.PointerString(m)
}

// VnicAttachmentLifecycleStateEnum Enum with underlying type: string
type VnicAttachmentLifecycleStateEnum string

// Set of constants representing the allowable values for VnicAttachmentLifecycleStateEnum
const (
	VnicAttachmentLifecycleStateAttaching VnicAttachmentLifecycleStateEnum = "ATTACHING"
	VnicAttachmentLifecycleStateAttached  VnicAttachmentLifecycleStateEnum = "ATTACHED"
	VnicAttachmentLifecycleStateDetaching VnicAttachmentLifecycleStateEnum = "DETACHING"
	VnicAttachmentLifecycleStateDetached  VnicAttachmentLifecycleStateEnum = "DETACHED"
)

var mappingVnicAttachmentLifecycleState = map[string]VnicAttachmentLifecycleStateEnum{
	"ATTACHING": VnicAttachmentLifecycleStateAttaching,
	"ATTACHED":  VnicAttachmentLifecycleStateAttached,
	"DETACHING": VnicAttachmentLifecycleStateDetaching,
	"DETACHED":  VnicAttachmentLifecycleStateDetached,
}

// GetVnicAttachmentLifecycleStateEnumValues Enumerates the set of values for VnicAttachmentLifecycleStateEnum
func GetVnicAttachmentLifecycleStateEnumValues() []VnicAttachmentLifecycleStateEnum {
	values := make([]VnicAttachmentLifecycleStateEnum, 0)
	for _, v := range mappingVnicAttachmentLifecycleState {
		values = append(values, v)
	}
	return values
}
