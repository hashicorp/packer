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

// FastConnectProviderServiceKey A provider service key and its details. A provider service key is an identifier for a provider's
// virtual circuit.
type FastConnectProviderServiceKey struct {

	// The service key that the provider gives you when you set up a virtual circuit connection
	// from the provider to Oracle Cloud Infrastructure. Use this value as the `providerServiceKeyName`
	// query parameter for
	// GetFastConnectProviderServiceKey.
	Name *string `mandatory:"true" json:"name"`

	// The provisioned data rate of the connection. To get a list of the
	// available bandwidth levels (that is, shapes), see
	// ListFastConnectProviderVirtualCircuitBandwidthShapes.
	// Example: `10 Gbps`
	BandwidthShapeName *string `mandatory:"false" json:"bandwidthShapeName"`

	// The provider's peering location.
	PeeringLocation *string `mandatory:"false" json:"peeringLocation"`
}

func (m FastConnectProviderServiceKey) String() string {
	return common.PointerString(m)
}
