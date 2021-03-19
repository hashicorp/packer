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

// LaunchInstancePlatformConfig The platform configuration requested for the instance.
// If the parameter is provided, the instance is created with the platform configured as specified. If some
// properties are missing or the entire parameter is not provided, the instance is created
// with the default configuration values for the `shape` that you specify.
// Each shape only supports certain configurable values. If the values that you provide are not valid for the
// specified `shape`, an error is returned.
type LaunchInstancePlatformConfig interface {
}

type launchinstanceplatformconfig struct {
	JsonData []byte
	Type     string `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *launchinstanceplatformconfig) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerlaunchinstanceplatformconfig launchinstanceplatformconfig
	s := struct {
		Model Unmarshalerlaunchinstanceplatformconfig
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Type = s.Model.Type

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *launchinstanceplatformconfig) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "AMD_MILAN_BM":
		mm := AmdMilanBmLaunchInstancePlatformConfig{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

func (m launchinstanceplatformconfig) String() string {
	return common.PointerString(m)
}

// LaunchInstancePlatformConfigTypeEnum Enum with underlying type: string
type LaunchInstancePlatformConfigTypeEnum string

// Set of constants representing the allowable values for LaunchInstancePlatformConfigTypeEnum
const (
	LaunchInstancePlatformConfigTypeAmdMilanBm LaunchInstancePlatformConfigTypeEnum = "AMD_MILAN_BM"
)

var mappingLaunchInstancePlatformConfigType = map[string]LaunchInstancePlatformConfigTypeEnum{
	"AMD_MILAN_BM": LaunchInstancePlatformConfigTypeAmdMilanBm,
}

// GetLaunchInstancePlatformConfigTypeEnumValues Enumerates the set of values for LaunchInstancePlatformConfigTypeEnum
func GetLaunchInstancePlatformConfigTypeEnumValues() []LaunchInstancePlatformConfigTypeEnum {
	values := make([]LaunchInstancePlatformConfigTypeEnum, 0)
	for _, v := range mappingLaunchInstancePlatformConfigType {
		values = append(values, v)
	}
	return values
}
