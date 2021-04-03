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
	"encoding/json"
	"github.com/oracle/oci-go-sdk/v36/common"
)

// AttachIScsiVolumeDetails The representation of AttachIScsiVolumeDetails
type AttachIScsiVolumeDetails struct {

	// The OCID of the instance.
	InstanceId *string `mandatory:"true" json:"instanceId"`

	// The OCID of the volume.
	VolumeId *string `mandatory:"true" json:"volumeId"`

	// The device name.
	Device *string `mandatory:"false" json:"device"`

	// A user-friendly name. Does not have to be unique, and it cannot be changed. Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Whether the attachment was created in read-only mode.
	IsReadOnly *bool `mandatory:"false" json:"isReadOnly"`

	// Whether the attachment should be created in shareable mode. If an attachment
	// is created in shareable mode, then other instances can attach the same volume, provided
	// that they also create their attachments in shareable mode. Only certain volume types can
	// be attached in shareable mode. Defaults to false if not specified.
	IsShareable *bool `mandatory:"false" json:"isShareable"`

	// Whether to use CHAP authentication for the volume attachment. Defaults to false.
	UseChap *bool `mandatory:"false" json:"useChap"`
}

//GetDevice returns Device
func (m AttachIScsiVolumeDetails) GetDevice() *string {
	return m.Device
}

//GetDisplayName returns DisplayName
func (m AttachIScsiVolumeDetails) GetDisplayName() *string {
	return m.DisplayName
}

//GetInstanceId returns InstanceId
func (m AttachIScsiVolumeDetails) GetInstanceId() *string {
	return m.InstanceId
}

//GetIsReadOnly returns IsReadOnly
func (m AttachIScsiVolumeDetails) GetIsReadOnly() *bool {
	return m.IsReadOnly
}

//GetIsShareable returns IsShareable
func (m AttachIScsiVolumeDetails) GetIsShareable() *bool {
	return m.IsShareable
}

//GetVolumeId returns VolumeId
func (m AttachIScsiVolumeDetails) GetVolumeId() *string {
	return m.VolumeId
}

func (m AttachIScsiVolumeDetails) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m AttachIScsiVolumeDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeAttachIScsiVolumeDetails AttachIScsiVolumeDetails
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeAttachIScsiVolumeDetails
	}{
		"iscsi",
		(MarshalTypeAttachIScsiVolumeDetails)(m),
	}

	return json.Marshal(&s)
}
