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

// AddSecurityRuleDetails A rule for allowing inbound (INGRESS) or outbound (EGRESS) IP packets.
type AddSecurityRuleDetails struct {

	// Direction of the security rule. Set to `EGRESS` for rules to allow outbound IP packets,
	// or `INGRESS` for rules to allow inbound IP packets.
	Direction AddSecurityRuleDetailsDirectionEnum `mandatory:"true" json:"direction"`

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
	DestinationType AddSecurityRuleDetailsDestinationTypeEnum `mandatory:"false" json:"destinationType,omitempty"`

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
	SourceType AddSecurityRuleDetailsSourceTypeEnum `mandatory:"false" json:"sourceType,omitempty"`

	TcpOptions *TcpOptions `mandatory:"false" json:"tcpOptions"`

	UdpOptions *UdpOptions `mandatory:"false" json:"udpOptions"`
}

func (m AddSecurityRuleDetails) String() string {
	return common.PointerString(m)
}

// AddSecurityRuleDetailsDestinationTypeEnum Enum with underlying type: string
type AddSecurityRuleDetailsDestinationTypeEnum string

// Set of constants representing the allowable values for AddSecurityRuleDetailsDestinationTypeEnum
const (
	AddSecurityRuleDetailsDestinationTypeCidrBlock            AddSecurityRuleDetailsDestinationTypeEnum = "CIDR_BLOCK"
	AddSecurityRuleDetailsDestinationTypeServiceCidrBlock     AddSecurityRuleDetailsDestinationTypeEnum = "SERVICE_CIDR_BLOCK"
	AddSecurityRuleDetailsDestinationTypeNetworkSecurityGroup AddSecurityRuleDetailsDestinationTypeEnum = "NETWORK_SECURITY_GROUP"
)

var mappingAddSecurityRuleDetailsDestinationType = map[string]AddSecurityRuleDetailsDestinationTypeEnum{
	"CIDR_BLOCK":             AddSecurityRuleDetailsDestinationTypeCidrBlock,
	"SERVICE_CIDR_BLOCK":     AddSecurityRuleDetailsDestinationTypeServiceCidrBlock,
	"NETWORK_SECURITY_GROUP": AddSecurityRuleDetailsDestinationTypeNetworkSecurityGroup,
}

// GetAddSecurityRuleDetailsDestinationTypeEnumValues Enumerates the set of values for AddSecurityRuleDetailsDestinationTypeEnum
func GetAddSecurityRuleDetailsDestinationTypeEnumValues() []AddSecurityRuleDetailsDestinationTypeEnum {
	values := make([]AddSecurityRuleDetailsDestinationTypeEnum, 0)
	for _, v := range mappingAddSecurityRuleDetailsDestinationType {
		values = append(values, v)
	}
	return values
}

// AddSecurityRuleDetailsDirectionEnum Enum with underlying type: string
type AddSecurityRuleDetailsDirectionEnum string

// Set of constants representing the allowable values for AddSecurityRuleDetailsDirectionEnum
const (
	AddSecurityRuleDetailsDirectionEgress  AddSecurityRuleDetailsDirectionEnum = "EGRESS"
	AddSecurityRuleDetailsDirectionIngress AddSecurityRuleDetailsDirectionEnum = "INGRESS"
)

var mappingAddSecurityRuleDetailsDirection = map[string]AddSecurityRuleDetailsDirectionEnum{
	"EGRESS":  AddSecurityRuleDetailsDirectionEgress,
	"INGRESS": AddSecurityRuleDetailsDirectionIngress,
}

// GetAddSecurityRuleDetailsDirectionEnumValues Enumerates the set of values for AddSecurityRuleDetailsDirectionEnum
func GetAddSecurityRuleDetailsDirectionEnumValues() []AddSecurityRuleDetailsDirectionEnum {
	values := make([]AddSecurityRuleDetailsDirectionEnum, 0)
	for _, v := range mappingAddSecurityRuleDetailsDirection {
		values = append(values, v)
	}
	return values
}

// AddSecurityRuleDetailsSourceTypeEnum Enum with underlying type: string
type AddSecurityRuleDetailsSourceTypeEnum string

// Set of constants representing the allowable values for AddSecurityRuleDetailsSourceTypeEnum
const (
	AddSecurityRuleDetailsSourceTypeCidrBlock            AddSecurityRuleDetailsSourceTypeEnum = "CIDR_BLOCK"
	AddSecurityRuleDetailsSourceTypeServiceCidrBlock     AddSecurityRuleDetailsSourceTypeEnum = "SERVICE_CIDR_BLOCK"
	AddSecurityRuleDetailsSourceTypeNetworkSecurityGroup AddSecurityRuleDetailsSourceTypeEnum = "NETWORK_SECURITY_GROUP"
)

var mappingAddSecurityRuleDetailsSourceType = map[string]AddSecurityRuleDetailsSourceTypeEnum{
	"CIDR_BLOCK":             AddSecurityRuleDetailsSourceTypeCidrBlock,
	"SERVICE_CIDR_BLOCK":     AddSecurityRuleDetailsSourceTypeServiceCidrBlock,
	"NETWORK_SECURITY_GROUP": AddSecurityRuleDetailsSourceTypeNetworkSecurityGroup,
}

// GetAddSecurityRuleDetailsSourceTypeEnumValues Enumerates the set of values for AddSecurityRuleDetailsSourceTypeEnum
func GetAddSecurityRuleDetailsSourceTypeEnumValues() []AddSecurityRuleDetailsSourceTypeEnum {
	values := make([]AddSecurityRuleDetailsSourceTypeEnum, 0)
	for _, v := range mappingAddSecurityRuleDetailsSourceType {
		values = append(values, v)
	}
	return values
}
