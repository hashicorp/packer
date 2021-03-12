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

// Shape A compute instance shape that can be used in LaunchInstance.
// For more information, see Overview of the Compute Service (https://docs.cloud.oracle.com/Content/Compute/Concepts/computeoverview.htm) and
// Compute Shapes (https://docs.cloud.oracle.com/Content/Compute/References/computeshapes.htm).
type Shape struct {

	// The name of the shape. You can enumerate all available shapes by calling
	// ListShapes.
	Shape *string `mandatory:"true" json:"shape"`

	// A short description of the shape's processor (CPU).
	ProcessorDescription *string `mandatory:"false" json:"processorDescription"`

	// The default number of OCPUs available for this shape.
	Ocpus *float32 `mandatory:"false" json:"ocpus"`

	// The default amount of memory available for this shape, in gigabytes.
	MemoryInGBs *float32 `mandatory:"false" json:"memoryInGBs"`

	// The networking bandwidth available for this shape, in gigabits per second.
	NetworkingBandwidthInGbps *float32 `mandatory:"false" json:"networkingBandwidthInGbps"`

	// The maximum number of VNIC attachments available for this shape.
	MaxVnicAttachments *int `mandatory:"false" json:"maxVnicAttachments"`

	// The number of GPUs available for this shape.
	Gpus *int `mandatory:"false" json:"gpus"`

	// A short description of the graphics processing unit (GPU) available for this shape.
	// If the shape does not have any GPUs, this field is `null`.
	GpuDescription *string `mandatory:"false" json:"gpuDescription"`

	// The number of local disks available for this shape.
	LocalDisks *int `mandatory:"false" json:"localDisks"`

	// The aggregate size of the local disks available for this shape, in gigabytes.
	// If the shape does not have any local disks, this field is `null`.
	LocalDisksTotalSizeInGBs *float32 `mandatory:"false" json:"localDisksTotalSizeInGBs"`

	// A short description of the local disks available for this shape.
	// If the shape does not have any local disks, this field is `null`.
	LocalDiskDescription *string `mandatory:"false" json:"localDiskDescription"`

	OcpuOptions *ShapeOcpuOptions `mandatory:"false" json:"ocpuOptions"`

	MemoryOptions *ShapeMemoryOptions `mandatory:"false" json:"memoryOptions"`

	NetworkingBandwidthOptions *ShapeNetworkingBandwidthOptions `mandatory:"false" json:"networkingBandwidthOptions"`

	MaxVnicAttachmentOptions *ShapeMaxVnicAttachmentOptions `mandatory:"false" json:"maxVnicAttachmentOptions"`
}

func (m Shape) String() string {
	return common.PointerString(m)
}
