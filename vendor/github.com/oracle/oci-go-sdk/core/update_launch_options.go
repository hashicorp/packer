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

// UpdateLaunchOptions Options for tuning the compatibility and performance of VM shapes.
type UpdateLaunchOptions struct {

	// Emulation type for the boot volume.
	// * `ISCSI` - ISCSI attached block storage device.
	// * `PARAVIRTUALIZED` - Paravirtualized disk. This is the default for boot volumes and remote block
	// storage volumes on Oracle-provided plaform images.
	// Before you change the boot volume attachment type, detach all block volumes and VNICs except for
	// the boot volume and the primary VNIC.
	// If the instance is running when you change the boot volume attachment type, it will be rebooted.
	// **Note:** Some instances might not function properly if you change the boot volume attachment type. After
	// the instance reboots and is running, connect to it. If the connection fails or the OS doesn't behave
	// as expected, the changes are not supported. Revert the instance to the original boot volume attachment type.
	BootVolumeType UpdateLaunchOptionsBootVolumeTypeEnum `mandatory:"false" json:"bootVolumeType,omitempty"`

	// Emulation type for the physical network interface card (NIC).
	// * `VFIO` - Direct attached Virtual Function network controller. This is the networking type
	// when you launch an instance using hardware-assisted (SR-IOV) networking.
	// * `PARAVIRTUALIZED` - VM instances launch with paravirtualized devices using VirtIO drivers.
	// Before you change the networking type, detach all VNICs and block volumes except for the primary
	// VNIC and the boot volume.
	// The image must have paravirtualized drivers installed. For more information, see
	// Editing an Instance (https://docs.cloud.oracle.com/iaas/Content/Compute/Tasks/resizinginstances.htm).
	// If the instance is running when you change the network type, it will be rebooted.
	// **Note:** Some instances might not function properly if you change the networking type. After
	// the instance reboots and is running, connect to it. If the connection fails or the OS doesn't behave
	// as expected, the changes are not supported. Revert the instance to the original networking type.
	NetworkType UpdateLaunchOptionsNetworkTypeEnum `mandatory:"false" json:"networkType,omitempty"`

	// Whether to enable in-transit encryption for the boot volume's paravirtualized attachment.
	// Data in transit is transferred over an internal and highly secure network. If you have specific
	// compliance requirements related to the encryption of the data while it is moving between the
	// instance and the boot volume, you can enable in-transit encryption. In-transit encryption is
	// not enabled by default.
	// All boot volumes are encrypted at rest.
	// For more information, see Block Volume Encryption (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm#Encrypti).
	IsPvEncryptionInTransitEnabled *bool `mandatory:"false" json:"isPvEncryptionInTransitEnabled"`
}

func (m UpdateLaunchOptions) String() string {
	return common.PointerString(m)
}

// UpdateLaunchOptionsBootVolumeTypeEnum Enum with underlying type: string
type UpdateLaunchOptionsBootVolumeTypeEnum string

// Set of constants representing the allowable values for UpdateLaunchOptionsBootVolumeTypeEnum
const (
	UpdateLaunchOptionsBootVolumeTypeIscsi           UpdateLaunchOptionsBootVolumeTypeEnum = "ISCSI"
	UpdateLaunchOptionsBootVolumeTypeParavirtualized UpdateLaunchOptionsBootVolumeTypeEnum = "PARAVIRTUALIZED"
)

var mappingUpdateLaunchOptionsBootVolumeType = map[string]UpdateLaunchOptionsBootVolumeTypeEnum{
	"ISCSI":           UpdateLaunchOptionsBootVolumeTypeIscsi,
	"PARAVIRTUALIZED": UpdateLaunchOptionsBootVolumeTypeParavirtualized,
}

// GetUpdateLaunchOptionsBootVolumeTypeEnumValues Enumerates the set of values for UpdateLaunchOptionsBootVolumeTypeEnum
func GetUpdateLaunchOptionsBootVolumeTypeEnumValues() []UpdateLaunchOptionsBootVolumeTypeEnum {
	values := make([]UpdateLaunchOptionsBootVolumeTypeEnum, 0)
	for _, v := range mappingUpdateLaunchOptionsBootVolumeType {
		values = append(values, v)
	}
	return values
}

// UpdateLaunchOptionsNetworkTypeEnum Enum with underlying type: string
type UpdateLaunchOptionsNetworkTypeEnum string

// Set of constants representing the allowable values for UpdateLaunchOptionsNetworkTypeEnum
const (
	UpdateLaunchOptionsNetworkTypeVfio            UpdateLaunchOptionsNetworkTypeEnum = "VFIO"
	UpdateLaunchOptionsNetworkTypeParavirtualized UpdateLaunchOptionsNetworkTypeEnum = "PARAVIRTUALIZED"
)

var mappingUpdateLaunchOptionsNetworkType = map[string]UpdateLaunchOptionsNetworkTypeEnum{
	"VFIO":            UpdateLaunchOptionsNetworkTypeVfio,
	"PARAVIRTUALIZED": UpdateLaunchOptionsNetworkTypeParavirtualized,
}

// GetUpdateLaunchOptionsNetworkTypeEnumValues Enumerates the set of values for UpdateLaunchOptionsNetworkTypeEnum
func GetUpdateLaunchOptionsNetworkTypeEnumValues() []UpdateLaunchOptionsNetworkTypeEnum {
	values := make([]UpdateLaunchOptionsNetworkTypeEnum, 0)
	for _, v := range mappingUpdateLaunchOptionsNetworkType {
		values = append(values, v)
	}
	return values
}
