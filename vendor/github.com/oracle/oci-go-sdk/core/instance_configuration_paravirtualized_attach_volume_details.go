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

// InstanceConfigurationParavirtualizedAttachVolumeDetails The representation of InstanceConfigurationParavirtualizedAttachVolumeDetails
type InstanceConfigurationParavirtualizedAttachVolumeDetails struct {

	// A user-friendly name. Does not have to be unique, and it cannot be changed. Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Whether the attachment should be created in read-only mode.
	IsReadOnly *bool `mandatory:"false" json:"isReadOnly"`

	// The device name.
	Device *string `mandatory:"false" json:"device"`

	// Whether the attachment should be created in shareable mode. If an attachment
	// is created in shareable mode, then other instances can attach the same volume, provided
	// that they also create their attachments in shareable mode. Only certain volume types can
	// be attached in shareable mode. Defaults to false if not specified.
	IsShareable *bool `mandatory:"false" json:"isShareable"`

	// Whether to enable in-transit encryption for the data volume's paravirtualized attachment. The default value is false.
	IsPvEncryptionInTransitEnabled *bool `mandatory:"false" json:"isPvEncryptionInTransitEnabled"`
}

//GetDisplayName returns DisplayName
func (m InstanceConfigurationParavirtualizedAttachVolumeDetails) GetDisplayName() *string {
	return m.DisplayName
}

//GetIsReadOnly returns IsReadOnly
func (m InstanceConfigurationParavirtualizedAttachVolumeDetails) GetIsReadOnly() *bool {
	return m.IsReadOnly
}

//GetDevice returns Device
func (m InstanceConfigurationParavirtualizedAttachVolumeDetails) GetDevice() *string {
	return m.Device
}

//GetIsShareable returns IsShareable
func (m InstanceConfigurationParavirtualizedAttachVolumeDetails) GetIsShareable() *bool {
	return m.IsShareable
}

func (m InstanceConfigurationParavirtualizedAttachVolumeDetails) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m InstanceConfigurationParavirtualizedAttachVolumeDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeInstanceConfigurationParavirtualizedAttachVolumeDetails InstanceConfigurationParavirtualizedAttachVolumeDetails
	s := struct {
		DiscriminatorParam string `json:"type"`
		MarshalTypeInstanceConfigurationParavirtualizedAttachVolumeDetails
	}{
		"paravirtualized",
		(MarshalTypeInstanceConfigurationParavirtualizedAttachVolumeDetails)(m),
	}

	return json.Marshal(&s)
}
