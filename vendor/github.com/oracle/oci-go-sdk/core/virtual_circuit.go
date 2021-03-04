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

// VirtualCircuit For use with Oracle Cloud Infrastructure FastConnect.
// A virtual circuit is an isolated network path that runs over one or more physical
// network connections to provide a single, logical connection between the edge router
// on the customer's existing network and Oracle Cloud Infrastructure. *Private*
// virtual circuits support private peering, and *public* virtual circuits support
// public peering. For more information, see FastConnect Overview (https://docs.cloud.oracle.com/Content/Network/Concepts/fastconnect.htm).
// Each virtual circuit is made up of information shared between a customer, Oracle,
// and a provider (if the customer is using FastConnect via a provider). Who fills in
// a given property of a virtual circuit depends on whether the BGP session related to
// that virtual circuit goes from the customer's edge router to Oracle, or from the provider's
// edge router to Oracle. Also, in the case where the customer is using a provider, values
// for some of the properties may not be present immediately, but may get filled in as the
// provider and Oracle each do their part to provision the virtual circuit.
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.cloud.oracle.com/Content/Identity/Concepts/policygetstarted.htm).
// **Warning:** Oracle recommends that you avoid using any confidential information when you
// supply string values using the API.
type VirtualCircuit struct {

	// The provisioned data rate of the connection.  To get a list of the
	// available bandwidth levels (that is, shapes), see
	// ListFastConnectProviderVirtualCircuitBandwidthShapes.
	// Example: `10 Gbps`
	BandwidthShapeName *string `mandatory:"false" json:"bandwidthShapeName"`

	// Deprecated. Instead use the information in
	// FastConnectProviderService.
	BgpManagement VirtualCircuitBgpManagementEnum `mandatory:"false" json:"bgpManagement,omitempty"`

	// The state of the BGP session associated with the virtual circuit.
	BgpSessionState VirtualCircuitBgpSessionStateEnum `mandatory:"false" json:"bgpSessionState,omitempty"`

	// The OCID of the compartment containing the virtual circuit.
	CompartmentId *string `mandatory:"false" json:"compartmentId"`

	// An array of mappings, each containing properties for a
	// cross-connect or cross-connect group that is associated with this
	// virtual circuit.
	CrossConnectMappings []CrossConnectMapping `mandatory:"false" json:"crossConnectMappings"`

	// Deprecated. Instead use `customerAsn`.
	// If you specify values for both, the request will be rejected.
	CustomerBgpAsn *int `mandatory:"false" json:"customerBgpAsn"`

	// The BGP ASN of the network at the other end of the BGP
	// session from Oracle. If the session is between the customer's
	// edge router and Oracle, the value is the customer's ASN. If the BGP
	// session is between the provider's edge router and Oracle, the value
	// is the provider's ASN.
	// Can be a 2-byte or 4-byte ASN. Uses "asplain" format.
	CustomerAsn *int64 `mandatory:"false" json:"customerAsn"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The OCID of the customer's Drg
	// that this virtual circuit uses. Applicable only to private virtual circuits.
	GatewayId *string `mandatory:"false" json:"gatewayId"`

	// The virtual circuit's Oracle ID (OCID).
	Id *string `mandatory:"false" json:"id"`

	// The virtual circuit's current state. For information about
	// the different states, see
	// FastConnect Overview (https://docs.cloud.oracle.com/Content/Network/Concepts/fastconnect.htm).
	LifecycleState VirtualCircuitLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`

	// The Oracle BGP ASN.
	OracleBgpAsn *int `mandatory:"false" json:"oracleBgpAsn"`

	// Deprecated. Instead use `providerServiceId`.
	ProviderName *string `mandatory:"false" json:"providerName"`

	// The OCID of the service offered by the provider (if the customer is connecting via a provider).
	ProviderServiceId *string `mandatory:"false" json:"providerServiceId"`

	// The service key name offered by the provider (if the customer is connecting via a provider).
	ProviderServiceKeyName *string `mandatory:"false" json:"providerServiceKeyName"`

	// Deprecated. Instead use `providerServiceId`.
	ProviderServiceName *string `mandatory:"false" json:"providerServiceName"`

	// The provider's state in relation to this virtual circuit (if the
	// customer is connecting via a provider). ACTIVE means
	// the provider has provisioned the virtual circuit from their end.
	// INACTIVE means the provider has not yet provisioned the virtual
	// circuit, or has de-provisioned it.
	ProviderState VirtualCircuitProviderStateEnum `mandatory:"false" json:"providerState,omitempty"`

	// For a public virtual circuit. The public IP prefixes (CIDRs) the customer wants to
	// advertise across the connection. All prefix sizes are allowed.
	PublicPrefixes []string `mandatory:"false" json:"publicPrefixes"`

	// Provider-supplied reference information about this virtual circuit
	// (if the customer is connecting via a provider).
	ReferenceComment *string `mandatory:"false" json:"referenceComment"`

	// The Oracle Cloud Infrastructure region where this virtual
	// circuit is located.
	Region *string `mandatory:"false" json:"region"`

	// Provider service type.
	ServiceType VirtualCircuitServiceTypeEnum `mandatory:"false" json:"serviceType,omitempty"`

	// The date and time the virtual circuit was created,
	// in the format defined by RFC3339 (https://tools.ietf.org/html/rfc3339).
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// Whether the virtual circuit supports private or public peering. For more information,
	// see FastConnect Overview (https://docs.cloud.oracle.com/Content/Network/Concepts/fastconnect.htm).
	Type VirtualCircuitTypeEnum `mandatory:"false" json:"type,omitempty"`
}

func (m VirtualCircuit) String() string {
	return common.PointerString(m)
}

// VirtualCircuitBgpManagementEnum Enum with underlying type: string
type VirtualCircuitBgpManagementEnum string

// Set of constants representing the allowable values for VirtualCircuitBgpManagementEnum
const (
	VirtualCircuitBgpManagementCustomerManaged VirtualCircuitBgpManagementEnum = "CUSTOMER_MANAGED"
	VirtualCircuitBgpManagementProviderManaged VirtualCircuitBgpManagementEnum = "PROVIDER_MANAGED"
	VirtualCircuitBgpManagementOracleManaged   VirtualCircuitBgpManagementEnum = "ORACLE_MANAGED"
)

var mappingVirtualCircuitBgpManagement = map[string]VirtualCircuitBgpManagementEnum{
	"CUSTOMER_MANAGED": VirtualCircuitBgpManagementCustomerManaged,
	"PROVIDER_MANAGED": VirtualCircuitBgpManagementProviderManaged,
	"ORACLE_MANAGED":   VirtualCircuitBgpManagementOracleManaged,
}

// GetVirtualCircuitBgpManagementEnumValues Enumerates the set of values for VirtualCircuitBgpManagementEnum
func GetVirtualCircuitBgpManagementEnumValues() []VirtualCircuitBgpManagementEnum {
	values := make([]VirtualCircuitBgpManagementEnum, 0)
	for _, v := range mappingVirtualCircuitBgpManagement {
		values = append(values, v)
	}
	return values
}

// VirtualCircuitBgpSessionStateEnum Enum with underlying type: string
type VirtualCircuitBgpSessionStateEnum string

// Set of constants representing the allowable values for VirtualCircuitBgpSessionStateEnum
const (
	VirtualCircuitBgpSessionStateUp   VirtualCircuitBgpSessionStateEnum = "UP"
	VirtualCircuitBgpSessionStateDown VirtualCircuitBgpSessionStateEnum = "DOWN"
)

var mappingVirtualCircuitBgpSessionState = map[string]VirtualCircuitBgpSessionStateEnum{
	"UP":   VirtualCircuitBgpSessionStateUp,
	"DOWN": VirtualCircuitBgpSessionStateDown,
}

// GetVirtualCircuitBgpSessionStateEnumValues Enumerates the set of values for VirtualCircuitBgpSessionStateEnum
func GetVirtualCircuitBgpSessionStateEnumValues() []VirtualCircuitBgpSessionStateEnum {
	values := make([]VirtualCircuitBgpSessionStateEnum, 0)
	for _, v := range mappingVirtualCircuitBgpSessionState {
		values = append(values, v)
	}
	return values
}

// VirtualCircuitLifecycleStateEnum Enum with underlying type: string
type VirtualCircuitLifecycleStateEnum string

// Set of constants representing the allowable values for VirtualCircuitLifecycleStateEnum
const (
	VirtualCircuitLifecycleStatePendingProvider VirtualCircuitLifecycleStateEnum = "PENDING_PROVIDER"
	VirtualCircuitLifecycleStateVerifying       VirtualCircuitLifecycleStateEnum = "VERIFYING"
	VirtualCircuitLifecycleStateProvisioning    VirtualCircuitLifecycleStateEnum = "PROVISIONING"
	VirtualCircuitLifecycleStateProvisioned     VirtualCircuitLifecycleStateEnum = "PROVISIONED"
	VirtualCircuitLifecycleStateFailed          VirtualCircuitLifecycleStateEnum = "FAILED"
	VirtualCircuitLifecycleStateInactive        VirtualCircuitLifecycleStateEnum = "INACTIVE"
	VirtualCircuitLifecycleStateTerminating     VirtualCircuitLifecycleStateEnum = "TERMINATING"
	VirtualCircuitLifecycleStateTerminated      VirtualCircuitLifecycleStateEnum = "TERMINATED"
)

var mappingVirtualCircuitLifecycleState = map[string]VirtualCircuitLifecycleStateEnum{
	"PENDING_PROVIDER": VirtualCircuitLifecycleStatePendingProvider,
	"VERIFYING":        VirtualCircuitLifecycleStateVerifying,
	"PROVISIONING":     VirtualCircuitLifecycleStateProvisioning,
	"PROVISIONED":      VirtualCircuitLifecycleStateProvisioned,
	"FAILED":           VirtualCircuitLifecycleStateFailed,
	"INACTIVE":         VirtualCircuitLifecycleStateInactive,
	"TERMINATING":      VirtualCircuitLifecycleStateTerminating,
	"TERMINATED":       VirtualCircuitLifecycleStateTerminated,
}

// GetVirtualCircuitLifecycleStateEnumValues Enumerates the set of values for VirtualCircuitLifecycleStateEnum
func GetVirtualCircuitLifecycleStateEnumValues() []VirtualCircuitLifecycleStateEnum {
	values := make([]VirtualCircuitLifecycleStateEnum, 0)
	for _, v := range mappingVirtualCircuitLifecycleState {
		values = append(values, v)
	}
	return values
}

// VirtualCircuitProviderStateEnum Enum with underlying type: string
type VirtualCircuitProviderStateEnum string

// Set of constants representing the allowable values for VirtualCircuitProviderStateEnum
const (
	VirtualCircuitProviderStateActive   VirtualCircuitProviderStateEnum = "ACTIVE"
	VirtualCircuitProviderStateInactive VirtualCircuitProviderStateEnum = "INACTIVE"
)

var mappingVirtualCircuitProviderState = map[string]VirtualCircuitProviderStateEnum{
	"ACTIVE":   VirtualCircuitProviderStateActive,
	"INACTIVE": VirtualCircuitProviderStateInactive,
}

// GetVirtualCircuitProviderStateEnumValues Enumerates the set of values for VirtualCircuitProviderStateEnum
func GetVirtualCircuitProviderStateEnumValues() []VirtualCircuitProviderStateEnum {
	values := make([]VirtualCircuitProviderStateEnum, 0)
	for _, v := range mappingVirtualCircuitProviderState {
		values = append(values, v)
	}
	return values
}

// VirtualCircuitServiceTypeEnum Enum with underlying type: string
type VirtualCircuitServiceTypeEnum string

// Set of constants representing the allowable values for VirtualCircuitServiceTypeEnum
const (
	VirtualCircuitServiceTypeColocated VirtualCircuitServiceTypeEnum = "COLOCATED"
	VirtualCircuitServiceTypeLayer2    VirtualCircuitServiceTypeEnum = "LAYER2"
	VirtualCircuitServiceTypeLayer3    VirtualCircuitServiceTypeEnum = "LAYER3"
)

var mappingVirtualCircuitServiceType = map[string]VirtualCircuitServiceTypeEnum{
	"COLOCATED": VirtualCircuitServiceTypeColocated,
	"LAYER2":    VirtualCircuitServiceTypeLayer2,
	"LAYER3":    VirtualCircuitServiceTypeLayer3,
}

// GetVirtualCircuitServiceTypeEnumValues Enumerates the set of values for VirtualCircuitServiceTypeEnum
func GetVirtualCircuitServiceTypeEnumValues() []VirtualCircuitServiceTypeEnum {
	values := make([]VirtualCircuitServiceTypeEnum, 0)
	for _, v := range mappingVirtualCircuitServiceType {
		values = append(values, v)
	}
	return values
}

// VirtualCircuitTypeEnum Enum with underlying type: string
type VirtualCircuitTypeEnum string

// Set of constants representing the allowable values for VirtualCircuitTypeEnum
const (
	VirtualCircuitTypePublic  VirtualCircuitTypeEnum = "PUBLIC"
	VirtualCircuitTypePrivate VirtualCircuitTypeEnum = "PRIVATE"
)

var mappingVirtualCircuitType = map[string]VirtualCircuitTypeEnum{
	"PUBLIC":  VirtualCircuitTypePublic,
	"PRIVATE": VirtualCircuitTypePrivate,
}

// GetVirtualCircuitTypeEnumValues Enumerates the set of values for VirtualCircuitTypeEnum
func GetVirtualCircuitTypeEnumValues() []VirtualCircuitTypeEnum {
	values := make([]VirtualCircuitTypeEnum, 0)
	for _, v := range mappingVirtualCircuitType {
		values = append(values, v)
	}
	return values
}
