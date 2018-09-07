// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Public DNS Service
//
// API for managing DNS zones, records, and policies.
//

package dns

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateZoneDetails The body for defining a new zone.
type CreateZoneDetails struct {

	// The name of the zone.
	Name *string `mandatory:"true" json:"name"`

	// The type of the zone. Must be either `PRIMARY` or `SECONDARY`.
	ZoneType CreateZoneDetailsZoneTypeEnum `mandatory:"true" json:"zoneType"`

	// The OCID of the compartment containing the zone.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// External master servers for the zone.
	ExternalMasters []ExternalMaster `mandatory:"false" json:"externalMasters"`
}

func (m CreateZoneDetails) String() string {
	return common.PointerString(m)
}

// CreateZoneDetailsZoneTypeEnum Enum with underlying type: string
type CreateZoneDetailsZoneTypeEnum string

// Set of constants representing the allowable values for CreateZoneDetailsZoneType
const (
	CreateZoneDetailsZoneTypePrimary   CreateZoneDetailsZoneTypeEnum = "PRIMARY"
	CreateZoneDetailsZoneTypeSecondary CreateZoneDetailsZoneTypeEnum = "SECONDARY"
)

var mappingCreateZoneDetailsZoneType = map[string]CreateZoneDetailsZoneTypeEnum{
	"PRIMARY":   CreateZoneDetailsZoneTypePrimary,
	"SECONDARY": CreateZoneDetailsZoneTypeSecondary,
}

// GetCreateZoneDetailsZoneTypeEnumValues Enumerates the set of values for CreateZoneDetailsZoneType
func GetCreateZoneDetailsZoneTypeEnumValues() []CreateZoneDetailsZoneTypeEnum {
	values := make([]CreateZoneDetailsZoneTypeEnum, 0)
	for _, v := range mappingCreateZoneDetailsZoneType {
		values = append(values, v)
	}
	return values
}
