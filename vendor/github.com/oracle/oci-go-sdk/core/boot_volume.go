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
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// BootVolume A detachable boot volume device that contains the image used to boot a Compute instance. For more information, see
// Overview of Boot Volumes (https://docs.cloud.oracle.com/Content/Block/Concepts/bootvolumes.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/Content/Identity/Concepts/policygetstarted.htm).
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
type BootVolume struct {

	// The availability domain of the boot volume.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment that contains the boot volume.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The boot volume's Oracle ID (OCID).
	Id *string `mandatory:"true" json:"id"`

	// The current state of a boot volume.
	LifecycleState BootVolumeLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The size of the volume in MBs. The value must be a multiple of 1024.
	// This field is deprecated. Please use sizeInGBs.
	SizeInMBs *int64 `mandatory:"true" json:"sizeInMBs"`

	// The date and time the boot volume was created. Format defined
	// by RFC3339 (https://tools.ietf.org/html/rfc3339).
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// System tags for this resource. Each key is predefined and scoped to a namespace.
	// Example: `{"foo-namespace": {"bar-key": "value"}}`
	SystemTags map[string]map[string]interface{} `mandatory:"false" json:"systemTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The image OCID used to create the boot volume.
	ImageId *string `mandatory:"false" json:"imageId"`

	// Specifies whether the boot volume's data has finished copying from the source boot volume or boot volume backup.
	IsHydrated *bool `mandatory:"false" json:"isHydrated"`

	// The number of volume performance units (VPUs) that will be applied to this boot volume per GB,
	// representing the Block Volume service's elastic performance options.
	// See Block Volume Elastic Performance (https://docs.cloud.oracle.com/Content/Block/Concepts/blockvolumeelasticperformance.htm) for more information.
	// Allowed values:
	//   * `10`: Represents Balanced option.
	//   * `20`: Represents Higher Performance option.
	VpusPerGB *int64 `mandatory:"false" json:"vpusPerGB"`

	// The size of the boot volume in GBs.
	SizeInGBs *int64 `mandatory:"false" json:"sizeInGBs"`

	// The boot volume source, either an existing boot volume in the same availability domain or a boot volume backup.
	// If null, this means that the boot volume was created from an image.
	SourceDetails BootVolumeSourceDetails `mandatory:"false" json:"sourceDetails"`

	// The OCID of the source volume group.
	VolumeGroupId *string `mandatory:"false" json:"volumeGroupId"`

	// The OCID of the Key Management master encryption key assigned to the boot volume.
	KmsKeyId *string `mandatory:"false" json:"kmsKeyId"`

	// Specifies whether the auto-tune performance is enabled for this boot volume.
	IsAutoTuneEnabled *bool `mandatory:"false" json:"isAutoTuneEnabled"`

	// The number of Volume Performance Units per GB that this boot volume is effectively tuned to when it's idle.
	AutoTunedVpusPerGB *int64 `mandatory:"false" json:"autoTunedVpusPerGB"`
}

func (m BootVolume) String() string {
	return common.PointerString(m)
}

// UnmarshalJSON unmarshals from json
func (m *BootVolume) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DefinedTags        map[string]map[string]interface{} `json:"definedTags"`
		SystemTags         map[string]map[string]interface{} `json:"systemTags"`
		DisplayName        *string                           `json:"displayName"`
		FreeformTags       map[string]string                 `json:"freeformTags"`
		ImageId            *string                           `json:"imageId"`
		IsHydrated         *bool                             `json:"isHydrated"`
		VpusPerGB          *int64                            `json:"vpusPerGB"`
		SizeInGBs          *int64                            `json:"sizeInGBs"`
		SourceDetails      bootvolumesourcedetails           `json:"sourceDetails"`
		VolumeGroupId      *string                           `json:"volumeGroupId"`
		KmsKeyId           *string                           `json:"kmsKeyId"`
		IsAutoTuneEnabled  *bool                             `json:"isAutoTuneEnabled"`
		AutoTunedVpusPerGB *int64                            `json:"autoTunedVpusPerGB"`
		AvailabilityDomain *string                           `json:"availabilityDomain"`
		CompartmentId      *string                           `json:"compartmentId"`
		Id                 *string                           `json:"id"`
		LifecycleState     BootVolumeLifecycleStateEnum      `json:"lifecycleState"`
		SizeInMBs          *int64                            `json:"sizeInMBs"`
		TimeCreated        *common.SDKTime                   `json:"timeCreated"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DefinedTags = model.DefinedTags

	m.SystemTags = model.SystemTags

	m.DisplayName = model.DisplayName

	m.FreeformTags = model.FreeformTags

	m.ImageId = model.ImageId

	m.IsHydrated = model.IsHydrated

	m.VpusPerGB = model.VpusPerGB

	m.SizeInGBs = model.SizeInGBs

	nn, e = model.SourceDetails.UnmarshalPolymorphicJSON(model.SourceDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.SourceDetails = nn.(BootVolumeSourceDetails)
	} else {
		m.SourceDetails = nil
	}

	m.VolumeGroupId = model.VolumeGroupId

	m.KmsKeyId = model.KmsKeyId

	m.IsAutoTuneEnabled = model.IsAutoTuneEnabled

	m.AutoTunedVpusPerGB = model.AutoTunedVpusPerGB

	m.AvailabilityDomain = model.AvailabilityDomain

	m.CompartmentId = model.CompartmentId

	m.Id = model.Id

	m.LifecycleState = model.LifecycleState

	m.SizeInMBs = model.SizeInMBs

	m.TimeCreated = model.TimeCreated

	return
}

// BootVolumeLifecycleStateEnum Enum with underlying type: string
type BootVolumeLifecycleStateEnum string

// Set of constants representing the allowable values for BootVolumeLifecycleStateEnum
const (
	BootVolumeLifecycleStateProvisioning BootVolumeLifecycleStateEnum = "PROVISIONING"
	BootVolumeLifecycleStateRestoring    BootVolumeLifecycleStateEnum = "RESTORING"
	BootVolumeLifecycleStateAvailable    BootVolumeLifecycleStateEnum = "AVAILABLE"
	BootVolumeLifecycleStateTerminating  BootVolumeLifecycleStateEnum = "TERMINATING"
	BootVolumeLifecycleStateTerminated   BootVolumeLifecycleStateEnum = "TERMINATED"
	BootVolumeLifecycleStateFaulty       BootVolumeLifecycleStateEnum = "FAULTY"
)

var mappingBootVolumeLifecycleState = map[string]BootVolumeLifecycleStateEnum{
	"PROVISIONING": BootVolumeLifecycleStateProvisioning,
	"RESTORING":    BootVolumeLifecycleStateRestoring,
	"AVAILABLE":    BootVolumeLifecycleStateAvailable,
	"TERMINATING":  BootVolumeLifecycleStateTerminating,
	"TERMINATED":   BootVolumeLifecycleStateTerminated,
	"FAULTY":       BootVolumeLifecycleStateFaulty,
}

// GetBootVolumeLifecycleStateEnumValues Enumerates the set of values for BootVolumeLifecycleStateEnum
func GetBootVolumeLifecycleStateEnumValues() []BootVolumeLifecycleStateEnum {
	values := make([]BootVolumeLifecycleStateEnum, 0)
	for _, v := range mappingBootVolumeLifecycleState {
		values = append(values, v)
	}
	return values
}
