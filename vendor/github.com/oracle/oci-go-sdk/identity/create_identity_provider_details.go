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

// CreateIdentityProviderDetails The representation of CreateIdentityProviderDetails
type CreateIdentityProviderDetails interface {

	// The OCID of your tenancy.
	GetCompartmentId() *string

	// The name you assign to the `IdentityProvider` during creation.
	// The name must be unique across all `IdentityProvider` objects in the
	// tenancy and cannot be changed.
	GetName() *string

	// The description you assign to the `IdentityProvider` during creation.
	// Does not have to be unique, and it's changeable.
	GetDescription() *string

	// The identity provider service or product.
	// Supported identity providers are Oracle Identity Cloud Service (IDCS) and Microsoft
	// Active Directory Federation Services (ADFS).
	// Example: `IDCS`
	GetProductType() CreateIdentityProviderDetailsProductTypeEnum
}

type createidentityproviderdetails struct {
	JsonData      []byte
	CompartmentId *string                                      `mandatory:"true" json:"compartmentId"`
	Name          *string                                      `mandatory:"true" json:"name"`
	Description   *string                                      `mandatory:"true" json:"description"`
	ProductType   CreateIdentityProviderDetailsProductTypeEnum `mandatory:"true" json:"productType"`
	Protocol      string                                       `json:"protocol"`
}

// UnmarshalJSON unmarshals json
func (m *createidentityproviderdetails) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalercreateidentityproviderdetails createidentityproviderdetails
	s := struct {
		Model Unmarshalercreateidentityproviderdetails
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.CompartmentId = s.Model.CompartmentId
	m.Name = s.Model.Name
	m.Description = s.Model.Description
	m.ProductType = s.Model.ProductType
	m.Protocol = s.Model.Protocol

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *createidentityproviderdetails) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {
	var err error
	switch m.Protocol {
	case "SAML2":
		mm := CreateSaml2IdentityProviderDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return m, nil
	}
}

//GetCompartmentId returns CompartmentId
func (m createidentityproviderdetails) GetCompartmentId() *string {
	return m.CompartmentId
}

//GetName returns Name
func (m createidentityproviderdetails) GetName() *string {
	return m.Name
}

//GetDescription returns Description
func (m createidentityproviderdetails) GetDescription() *string {
	return m.Description
}

//GetProductType returns ProductType
func (m createidentityproviderdetails) GetProductType() CreateIdentityProviderDetailsProductTypeEnum {
	return m.ProductType
}

func (m createidentityproviderdetails) String() string {
	return common.PointerString(m)
}

// CreateIdentityProviderDetailsProductTypeEnum Enum with underlying type: string
type CreateIdentityProviderDetailsProductTypeEnum string

// Set of constants representing the allowable values for CreateIdentityProviderDetailsProductType
const (
	CreateIdentityProviderDetailsProductTypeIdcs CreateIdentityProviderDetailsProductTypeEnum = "IDCS"
	CreateIdentityProviderDetailsProductTypeAdfs CreateIdentityProviderDetailsProductTypeEnum = "ADFS"
)

var mappingCreateIdentityProviderDetailsProductType = map[string]CreateIdentityProviderDetailsProductTypeEnum{
	"IDCS": CreateIdentityProviderDetailsProductTypeIdcs,
	"ADFS": CreateIdentityProviderDetailsProductTypeAdfs,
}

// GetCreateIdentityProviderDetailsProductTypeEnumValues Enumerates the set of values for CreateIdentityProviderDetailsProductType
func GetCreateIdentityProviderDetailsProductTypeEnumValues() []CreateIdentityProviderDetailsProductTypeEnum {
	values := make([]CreateIdentityProviderDetailsProductTypeEnum, 0)
	for _, v := range mappingCreateIdentityProviderDetailsProductType {
		values = append(values, v)
	}
	return values
}
