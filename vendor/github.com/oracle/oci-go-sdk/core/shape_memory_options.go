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

// ShapeMemoryOptions For a flexible shape, the amount of memory available for instances that use this shape.
// If this field is null, then this shape has a fixed amount of memory equivalent to `memoryInGBs`.
type ShapeMemoryOptions struct {

	// The minimum amount of memory, in gigabytes.
	MinInGBs *float32 `mandatory:"false" json:"minInGBs"`

	// The maximum amount of memory, in gigabytes.
	MaxInGBs *float32 `mandatory:"false" json:"maxInGBs"`

	// The default amount of memory per OCPU available for this shape, in gigabytes.
	DefaultPerOcpuInGBs *float32 `mandatory:"false" json:"defaultPerOcpuInGBs"`
}

func (m ShapeMemoryOptions) String() string {
	return common.PointerString(m)
}
