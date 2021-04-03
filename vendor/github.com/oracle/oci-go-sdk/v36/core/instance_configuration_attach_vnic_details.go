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

// InstanceConfigurationAttachVnicDetails The representation of InstanceConfigurationAttachVnicDetails
type InstanceConfigurationAttachVnicDetails struct {
	CreateVnicDetails *InstanceConfigurationCreateVnicDetails `mandatory:"false" json:"createVnicDetails"`

	// A user-friendly name for the attachment. Does not have to be unique, and it cannot be changed.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Which physical network interface card (NIC) the VNIC will use. Defaults to 0.
	// Certain bare metal instance shapes have two active physical NICs (0 and 1). If
	// you add a secondary VNIC to one of these instances, you can specify which NIC
	// the VNIC will use. For more information, see
	// Virtual Network Interface Cards (VNICs) (https://docs.cloud.oracle.com/Content/Network/Tasks/managingVNICs.htm).
	NicIndex *int `mandatory:"false" json:"nicIndex"`
}

func (m InstanceConfigurationAttachVnicDetails) String() string {
	return common.PointerString(m)
}
