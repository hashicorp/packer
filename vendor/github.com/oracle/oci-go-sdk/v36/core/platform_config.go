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
	"encoding/json"
	"github.com/oracle/oci-go-sdk/v36/common"
)

// PlatformConfig The platform configuration for the instance. The type of platform configuration is
// determined by the `type`.
type PlatformConfig interface {
}

type platformconfig struct {
	JsonData []byte
	Type     string `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *platformconfig) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerplatformconfig platformconfig
	s := struct {
		Model Unmarshalerplatformconfig
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Type = s.Model.Type

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *platformconfig) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "AMD_MILAN_BM":
		mm := AmdMilanBmPlatformConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

func (m platformconfig) String() string {
	return common.PointerString(m)
}

// PlatformConfigTypeEnum Enum with underlying type: string
type PlatformConfigTypeEnum string

// Set of constants representing the allowable values for PlatformConfigTypeEnum
const (
	PlatformConfigTypeAmdMilanBm PlatformConfigTypeEnum = "AMD_MILAN_BM"
)

var mappingPlatformConfigType = map[string]PlatformConfigTypeEnum{
	"AMD_MILAN_BM": PlatformConfigTypeAmdMilanBm,
}

// GetPlatformConfigTypeEnumValues Enumerates the set of values for PlatformConfigTypeEnum
func GetPlatformConfigTypeEnumValues() []PlatformConfigTypeEnum {
	values := make([]PlatformConfigTypeEnum, 0)
	for _, v := range mappingPlatformConfigType {
		values = append(values, v)
	}
	return values
}
