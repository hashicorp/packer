// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// CreateVolumeDetails The representation of CreateVolumeDetails
type CreateVolumeDetails struct {

	// The Availability Domain of the volume.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment that contains the volume.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The size of the volume in GBs.
	SizeInGBs *int `mandatory:"false" json:"sizeInGBs"`

	// The size of the volume in MBs. The value must be a multiple of 1024.
	// This field is deprecated. Use sizeInGBs instead.
	SizeInMBs *int `mandatory:"false" json:"sizeInMBs"`

	// Specifies the volume source details for a new Block volume. The volume source is either another Block volume in the same Availability Domain or a Block volume backup.
	// This is an optional field. If not specified or set to null, the new Block volume will be empty.
	// When specified, the new Block volume will contain data from the source volume or backup.
	SourceDetails VolumeSourceDetails `mandatory:"false" json:"sourceDetails"`

	// The OCID of the volume backup from which the data should be restored on the newly created volume.
	// This field is deprecated. Use the sourceDetails field instead to specify the
	// backup for the volume.
	VolumeBackupId *string `mandatory:"false" json:"volumeBackupId"`
}

func (m CreateVolumeDetails) String() string {
	return common.PointerString(m)
}

// UnmarshalJSON unmarshals from json
func (m *CreateVolumeDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DisplayName        *string             `json:"displayName"`
		SizeInGBs          *int                `json:"sizeInGBs"`
		SizeInMBs          *int                `json:"sizeInMBs"`
		SourceDetails      volumesourcedetails `json:"sourceDetails"`
		VolumeBackupId     *string             `json:"volumeBackupId"`
		AvailabilityDomain *string             `json:"availabilityDomain"`
		CompartmentId      *string             `json:"compartmentId"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	m.DisplayName = model.DisplayName
	m.SizeInGBs = model.SizeInGBs
	m.SizeInMBs = model.SizeInMBs
	nn, e := model.SourceDetails.UnmarshalPolymorphicJSON(model.SourceDetails.JsonData)
	if e != nil {
		return
	}
	m.SourceDetails = nn
	m.VolumeBackupId = model.VolumeBackupId
	m.AvailabilityDomain = model.AvailabilityDomain
	m.CompartmentId = model.CompartmentId
	return
}
