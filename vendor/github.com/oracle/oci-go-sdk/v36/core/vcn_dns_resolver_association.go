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

// VcnDnsResolverAssociation The information about the VCN and the DNS resolver in the association.
type VcnDnsResolverAssociation struct {

	// The OCID of the VCN in the association.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// The current state of the association.
	LifecycleState VcnDnsResolverAssociationLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID of the DNS resolver in the association.
	DnsResolverId *string `mandatory:"false" json:"dnsResolverId"`
}

func (m VcnDnsResolverAssociation) String() string {
	return common.PointerString(m)
}

// VcnDnsResolverAssociationLifecycleStateEnum Enum with underlying type: string
type VcnDnsResolverAssociationLifecycleStateEnum string

// Set of constants representing the allowable values for VcnDnsResolverAssociationLifecycleStateEnum
const (
	VcnDnsResolverAssociationLifecycleStateProvisioning VcnDnsResolverAssociationLifecycleStateEnum = "PROVISIONING"
	VcnDnsResolverAssociationLifecycleStateAvailable    VcnDnsResolverAssociationLifecycleStateEnum = "AVAILABLE"
	VcnDnsResolverAssociationLifecycleStateTerminating  VcnDnsResolverAssociationLifecycleStateEnum = "TERMINATING"
	VcnDnsResolverAssociationLifecycleStateTerminated   VcnDnsResolverAssociationLifecycleStateEnum = "TERMINATED"
)

var mappingVcnDnsResolverAssociationLifecycleState = map[string]VcnDnsResolverAssociationLifecycleStateEnum{
	"PROVISIONING": VcnDnsResolverAssociationLifecycleStateProvisioning,
	"AVAILABLE":    VcnDnsResolverAssociationLifecycleStateAvailable,
	"TERMINATING":  VcnDnsResolverAssociationLifecycleStateTerminating,
	"TERMINATED":   VcnDnsResolverAssociationLifecycleStateTerminated,
}

// GetVcnDnsResolverAssociationLifecycleStateEnumValues Enumerates the set of values for VcnDnsResolverAssociationLifecycleStateEnum
func GetVcnDnsResolverAssociationLifecycleStateEnumValues() []VcnDnsResolverAssociationLifecycleStateEnum {
	values := make([]VcnDnsResolverAssociationLifecycleStateEnum, 0)
	for _, v := range mappingVcnDnsResolverAssociationLifecycleState {
		values = append(values, v)
	}
	return values
}
