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

// UpdateSecurityRuleDetails A rule for allowing inbound (`direction`= INGRESS) or outbound (`direction`= EGRESS) IP packets.
type UpdateSecurityRuleDetails struct {

	// Direction of the security rule. Set to `EGRESS` for rules to allow outbound IP packets,
	// or `INGRESS` for rules to allow inbound IP packets.
	Direction UpdateSecurityRuleDetailsDirectionEnum `mandatory:"true" json:"direction"`

	// The Oracle-assigned ID of the security rule that you want to update. You can't change this value.
	// Example: `04ABEC`
	Id *string `mandatory:"true" json:"id"`

	// The transport protocol. Specify either `all` or an IPv4 protocol number as
	// defined in
	// Protocol Numbers (http://www.iana.org/assignments/protocol-numbers/protocol-numbers.xhtml).
	// Options are supported only for ICMP ("1"), TCP ("6"), UDP ("17"), and ICMPv6 ("58").
	Protocol *string `mandatory:"true" json:"protocol"`

	// An optional description of your choice for the rule. Avoid entering confidential information.
	Description *string `mandatory:"false" json:"description"`

	// Conceptually, this is the range of IP addresses that a packet originating from the instance
	// can go to.
	// Allowed values:
	//   * An IP address range in CIDR notation. For example: `192.168.1.0/24` or `2001:0db8:0123:45::/56`
	//     Note that IPv6 addressing is currently supported only in certain regions. See
	//     IPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	//   * The `cidrBlock` value for a Service, if you're
	//     setting up a security rule for traffic destined for a particular `Service` through
	//     a service gateway. For example: `oci-phx-objectstorage`.
	//   * The OCID of a NetworkSecurityGroup in the same
	//     VCN. The value can be the NSG that the rule belongs to if the rule's intent is to control
	//     traffic between VNICs in the same NSG.
	Destination *string `mandatory:"false" json:"destination"`

	// Type of destination for the rule. Required if `direction` = `EGRESS`.
	// Allowed values:
	//   * `CIDR_BLOCK`: If the rule's `destination` is an IP address range in CIDR notation.
	//   * `SERVICE_CIDR_BLOCK`: If the rule's `destination` is the `cidrBlock` value for a
	//     Service (the rule is for traffic destined for a
	//     particular `Service` through a service gateway).
	//   * `NETWORK_SECURITY_GROUP`: If the rule's `destination` is the OCID of a
	//     NetworkSecurityGroup.
	DestinationType UpdateSecurityRuleDetailsDestinationTypeEnum `mandatory:"false" json:"destinationType,omitempty"`

	IcmpOptions *IcmpOptions `mandatory:"false" json:"icmpOptions"`

	// A stateless rule allows traffic in one direction. Remember to add a corresponding
	// stateless rule in the other direction if you need to support bidirectional traffic. For
	// example, if egress traffic allows TCP destination port 80, there should be an ingress
	// rule to allow TCP source port 80. Defaults to false, which means the rule is stateful
	// and a corresponding rule is not necessary for bidirectional traffic.
	IsStateless *bool `mandatory:"false" json:"isStateless"`

	// Conceptually, this is the range of IP addresses that a packet coming into the instance
	// can come from.
	// Allowed values:
	//   * An IP address range in CIDR notation. For example: `192.168.1.0/24` or `2001:0db8:0123:45::/56`
	//     Note that IPv6 addressing is currently supported only in certain regions. See
	//     IPv6 Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/ipv6.htm).
	//   * The `cidrBlock` value for a Service, if you're
	//     setting up a security rule for traffic coming from a particular `Service` through
	//     a service gateway. For example: `oci-phx-objectstorage`.
	//   * The OCID of a NetworkSecurityGroup in the same
	//     VCN. The value can be the NSG that the rule belongs to if the rule's intent is to control
	//     traffic between VNICs in the same NSG.
	Source *string `mandatory:"false" json:"source"`

	// Type of source for the rule. Required if `direction` = `INGRESS`.
	//   * `CIDR_BLOCK`: If the rule's `source` is an IP address range in CIDR notation.
	//   * `SERVICE_CIDR_BLOCK`: If the rule's `source` is the `cidrBlock` value for a
	//     Service (the rule is for traffic coming from a
	//     particular `Service` through a service gateway).
	//   * `NETWORK_SECURITY_GROUP`: If the rule's `source` is the OCID of a
	//     NetworkSecurityGroup.
	SourceType UpdateSecurityRuleDetailsSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`

	TcpOptions *TcpOptions `mandatory:"false" json:"tcpOptions"`

	UdpOptions *UdpOptions `mandatory:"false" json:"udpOptions"`
}

func (m UpdateSecurityRuleDetails) String() string {
	return common.PointerString(m)
}

// UpdateSecurityRuleDetailsDestinationTypeEnum Enum with underlying type: string
type UpdateSecurityRuleDetailsDestinationTypeEnum string

// Set of constants representing the allowable values for UpdateSecurityRuleDetailsDestinationTypeEnum
const (
	UpdateSecurityRuleDetailsDestinationTypeCidrBlock            UpdateSecurityRuleDetailsDestinationTypeEnum = "CIDR_BLOCK"
	UpdateSecurityRuleDetailsDestinationTypeServiceCidrBlock     UpdateSecurityRuleDetailsDestinationTypeEnum = "SERVICE_CIDR_BLOCK"
	UpdateSecurityRuleDetailsDestinationTypeNetworkSecurityGroup UpdateSecurityRuleDetailsDestinationTypeEnum = "NETWORK_SECURITY_GROUP"
)

var mappingUpdateSecurityRuleDetailsDestinationType = map[string]UpdateSecurityRuleDetailsDestinationTypeEnum{
	"CIDR_BLOCK":             UpdateSecurityRuleDetailsDestinationTypeCidrBlock,
	"SERVICE_CIDR_BLOCK":     UpdateSecurityRuleDetailsDestinationTypeServiceCidrBlock,
	"NETWORK_SECURITY_GROUP": UpdateSecurityRuleDetailsDestinationTypeNetworkSecurityGroup,
}

// GetUpdateSecurityRuleDetailsDestinationTypeEnumValues Enumerates the set of values for UpdateSecurityRuleDetailsDestinationTypeEnum
func GetUpdateSecurityRuleDetailsDestinationTypeEnumValues() []UpdateSecurityRuleDetailsDestinationTypeEnum {
	values := make([]UpdateSecurityRuleDetailsDestinationTypeEnum, 0)
	for _, v := range mappingUpdateSecurityRuleDetailsDestinationType {
		values = append(values, v)
	}
	return values
}

// UpdateSecurityRuleDetailsDirectionEnum Enum with underlying type: string
type UpdateSecurityRuleDetailsDirectionEnum string

// Set of constants representing the allowable values for UpdateSecurityRuleDetailsDirectionEnum
const (
	UpdateSecurityRuleDetailsDirectionEgress  UpdateSecurityRuleDetailsDirectionEnum = "EGRESS"
	UpdateSecurityRuleDetailsDirectionIngress UpdateSecurityRuleDetailsDirectionEnum = "INGRESS"
)

var mappingUpdateSecurityRuleDetailsDirection = map[string]UpdateSecurityRuleDetailsDirectionEnum{
	"EGRESS":  UpdateSecurityRuleDetailsDirectionEgress,
	"INGRESS": UpdateSecurityRuleDetailsDirectionIngress,
}

// GetUpdateSecurityRuleDetailsDirectionEnumValues Enumerates the set of values for UpdateSecurityRuleDetailsDirectionEnum
func GetUpdateSecurityRuleDetailsDirectionEnumValues() []UpdateSecurityRuleDetailsDirectionEnum {
	values := make([]UpdateSecurityRuleDetailsDirectionEnum, 0)
	for _, v := range mappingUpdateSecurityRuleDetailsDirection {
		values = append(values, v)
	}
	return values
}

// UpdateSecurityRuleDetailsSourceTypeEnum Enum with underlying type: string
type UpdateSecurityRuleDetailsSourceTypeEnum string

// Set of constants representing the allowable values for UpdateSecurityRuleDetailsSourceTypeEnum
const (
	UpdateSecurityRuleDetailsSourceTypeCidrBlock            UpdateSecurityRuleDetailsSourceTypeEnum = "CIDR_BLOCK"
	UpdateSecurityRuleDetailsSourceTypeServiceCidrBlock     UpdateSecurityRuleDetailsSourceTypeEnum = "SERVICE_CIDR_BLOCK"
	UpdateSecurityRuleDetailsSourceTypeNetworkSecurityGroup UpdateSecurityRuleDetailsSourceTypeEnum = "NETWORK_SECURITY_GROUP"
)

var mappingUpdateSecurityRuleDetailsSourceType = map[string]UpdateSecurityRuleDetailsSourceTypeEnum{
	"CIDR_BLOCK":             UpdateSecurityRuleDetailsSourceTypeCidrBlock,
	"SERVICE_CIDR_BLOCK":     UpdateSecurityRuleDetailsSourceTypeServiceCidrBlock,
	"NETWORK_SECURITY_GROUP": UpdateSecurityRuleDetailsSourceTypeNetworkSecurityGroup,
}

// GetUpdateSecurityRuleDetailsSourceTypeEnumValues Enumerates the set of values for UpdateSecurityRuleDetailsSourceTypeEnum
func GetUpdateSecurityRuleDetailsSourceTypeEnumValues() []UpdateSecurityRuleDetailsSourceTypeEnum {
	values := make([]UpdateSecurityRuleDetailsSourceTypeEnum, 0)
	for _, v := range mappingUpdateSecurityRuleDetailsSourceType {
		values = append(values, v)
	}
	return values
}
