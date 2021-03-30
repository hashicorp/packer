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

// AddPublicIpPoolCapacityDetails The information used to add capacity to an IP pool.
type AddPublicIpPoolCapacityDetails struct {

	// The OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) of the `ByoipRange` resource to which the CIDR block belongs.
	ByoipRangeId *string `mandatory:"true" json:"byoipRangeId"`

	// The CIDR block to add to the public IP pool. It could be all of the CIDR block identified in `byoipRangeId`, or a subrange.
	// Example: `10.0.1.0/24`
	CidrBlock *string `mandatory:"true" json:"cidrBlock"`
}

func (m AddPublicIpPoolCapacityDetails) String() string {
	return common.PointerString(m)
}
