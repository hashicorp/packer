// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Identity and Access Management Service API
//
// APIs for managing users, groups, compartments, and policies.
//

package identity

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// IdentityProvider The resulting base object when you add an identity provider to your tenancy. A
// Saml2IdentityProvider
// is a specific type of `IdentityProvider` that supports the SAML 2.0 protocol. Each
// `IdentityProvider` object has its own OCID. For more information, see
// Identity Providers and Federation (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/federation.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access,
// see Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type IdentityProvider interface {

	// The OCID of the `IdentityProvider`.
	GetId() *string

	// The OCID of the tenancy containing the `IdentityProvider`.
	GetCompartmentId() *string

	// The name you assign to the `IdentityProvider` during creation. The name
	// must be unique across all `IdentityProvider` objects in the tenancy and
	// cannot be changed. This is the name federated users see when choosing
	// which identity provider to use when signing in to the Oracle Cloud Infrastructure
	// Console.
	GetName() *string

	// The description you assign to the `IdentityProvider` during creation. Does
	// not have to be unique, and it's changeable.
	GetDescription() *string

	// The identity provider service or product.
	// Supported identity providers are Oracle Identity Cloud Service (IDCS) and Microsoft
	// Active Directory Federation Services (ADFS).
	// Allowed values are:
	// - `ADFS`
	// - `IDCS`
	// Example: `IDCS`
	GetProductType() *string

	// Date and time the `IdentityProvider` was created, in the format defined by RFC3339.
	// Example: `2016-08-25T21:10:29.600Z`
	GetTimeCreated() *common.SDKTime

	// The current state. After creating an `IdentityProvider`, make sure its
	// `lifecycleState` changes from CREATING to ACTIVE before using it.
	GetLifecycleState() IdentityProviderLifecycleStateEnum

	// The detailed status of INACTIVE lifecycleState.
	GetInactiveStatus() *int
}

type identityprovider struct {
	JsonData       []byte
	Id             *string                            `mandatory:"true" json:"id"`
	CompartmentId  *string                            `mandatory:"true" json:"compartmentId"`
	Name           *string                            `mandatory:"true" json:"name"`
	Description    *string                            `mandatory:"true" json:"description"`
	ProductType    *string                            `mandatory:"true" json:"productType"`
	TimeCreated    *common.SDKTime                    `mandatory:"true" json:"timeCreated"`
	LifecycleState IdentityProviderLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`
	InactiveStatus *int                               `mandatory:"false" json:"inactiveStatus"`
	Protocol       string                             `json:"protocol"`
}

// UnmarshalJSON unmarshals json
func (m *identityprovider) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshaleridentityprovider identityprovider
	s := struct {
		Model Unmarshaleridentityprovider
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Id = s.Model.Id
	m.CompartmentId = s.Model.CompartmentId
	m.Name = s.Model.Name
	m.Description = s.Model.Description
	m.ProductType = s.Model.ProductType
	m.TimeCreated = s.Model.TimeCreated
	m.LifecycleState = s.Model.LifecycleState
	m.InactiveStatus = s.Model.InactiveStatus
	m.Protocol = s.Model.Protocol

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *identityprovider) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {
	var err error
	switch m.Protocol {
	case "SAML2":
		mm := Saml2IdentityProvider{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return m, nil
	}
}

//GetId returns Id
func (m identityprovider) GetId() *string {
	return m.Id
}

//GetCompartmentId returns CompartmentId
func (m identityprovider) GetCompartmentId() *string {
	return m.CompartmentId
}

//GetName returns Name
func (m identityprovider) GetName() *string {
	return m.Name
}

//GetDescription returns Description
func (m identityprovider) GetDescription() *string {
	return m.Description
}

//GetProductType returns ProductType
func (m identityprovider) GetProductType() *string {
	return m.ProductType
}

//GetTimeCreated returns TimeCreated
func (m identityprovider) GetTimeCreated() *common.SDKTime {
	return m.TimeCreated
}

//GetLifecycleState returns LifecycleState
func (m identityprovider) GetLifecycleState() IdentityProviderLifecycleStateEnum {
	return m.LifecycleState
}

//GetInactiveStatus returns InactiveStatus
func (m identityprovider) GetInactiveStatus() *int {
	return m.InactiveStatus
}

func (m identityprovider) String() string {
	return common.PointerString(m)
}

// IdentityProviderLifecycleStateEnum Enum with underlying type: string
type IdentityProviderLifecycleStateEnum string

// Set of constants representing the allowable values for IdentityProviderLifecycleState
const (
	IdentityProviderLifecycleStateCreating IdentityProviderLifecycleStateEnum = "CREATING"
	IdentityProviderLifecycleStateActive   IdentityProviderLifecycleStateEnum = "ACTIVE"
	IdentityProviderLifecycleStateInactive IdentityProviderLifecycleStateEnum = "INACTIVE"
	IdentityProviderLifecycleStateDeleting IdentityProviderLifecycleStateEnum = "DELETING"
	IdentityProviderLifecycleStateDeleted  IdentityProviderLifecycleStateEnum = "DELETED"
)

var mappingIdentityProviderLifecycleState = map[string]IdentityProviderLifecycleStateEnum{
	"CREATING": IdentityProviderLifecycleStateCreating,
	"ACTIVE":   IdentityProviderLifecycleStateActive,
	"INACTIVE": IdentityProviderLifecycleStateInactive,
	"DELETING": IdentityProviderLifecycleStateDeleting,
	"DELETED":  IdentityProviderLifecycleStateDeleted,
}

// GetIdentityProviderLifecycleStateEnumValues Enumerates the set of values for IdentityProviderLifecycleState
func GetIdentityProviderLifecycleStateEnumValues() []IdentityProviderLifecycleStateEnum {
	values := make([]IdentityProviderLifecycleStateEnum, 0)
	for _, v := range mappingIdentityProviderLifecycleState {
		values = append(values, v)
	}
	return values
}
