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

// ImageShapeCompatibilitySummary Summary information for a compatible image and shape.
type ImageShapeCompatibilitySummary struct {

	// The image OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm).
	ImageId *string `mandatory:"true" json:"imageId"`

	// The shape name.
	Shape *string `mandatory:"true" json:"shape"`

	OcpuConstraints *ImageOcpuConstraints `mandatory:"false" json:"ocpuConstraints"`
}

func (m ImageShapeCompatibilitySummary) String() string {
	return common.PointerString(m)
}
