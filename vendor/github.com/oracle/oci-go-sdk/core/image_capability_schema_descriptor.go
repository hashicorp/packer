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
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// ImageCapabilitySchemaDescriptor Image Capability Schema Descriptor is a type of capability for an image.
type ImageCapabilitySchemaDescriptor interface {
	GetSource() ImageCapabilitySchemaDescriptorSourceEnum
}

type imagecapabilityschemadescriptor struct {
	JsonData       []byte
	Source         ImageCapabilitySchemaDescriptorSourceEnum `mandatory:"true" json:"source"`
	DescriptorType string                                    `json:"descriptorType"`
}

// UnmarshalJSON unmarshals json
func (m *imagecapabilityschemadescriptor) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerimagecapabilityschemadescriptor imagecapabilityschemadescriptor
	s := struct {
		Model Unmarshalerimagecapabilityschemadescriptor
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Source = s.Model.Source
	m.DescriptorType = s.Model.DescriptorType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *imagecapabilityschemadescriptor) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.DescriptorType {
	case "enumstring":
		mm := EnumStringImageCapabilitySchemaDescriptor{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "enuminteger":
		mm := EnumIntegerImageCapabilityDescriptor{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "boolean":
		mm := BooleanImageCapabilitySchemaDescriptor{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

//GetSource returns Source
func (m imagecapabilityschemadescriptor) GetSource() ImageCapabilitySchemaDescriptorSourceEnum {
	return m.Source
}

func (m imagecapabilityschemadescriptor) String() string {
	return common.PointerString(m)
}

// ImageCapabilitySchemaDescriptorSourceEnum Enum with underlying type: string
type ImageCapabilitySchemaDescriptorSourceEnum string

// Set of constants representing the allowable values for ImageCapabilitySchemaDescriptorSourceEnum
const (
	ImageCapabilitySchemaDescriptorSourceGlobal ImageCapabilitySchemaDescriptorSourceEnum = "GLOBAL"
	ImageCapabilitySchemaDescriptorSourceImage  ImageCapabilitySchemaDescriptorSourceEnum = "IMAGE"
)

var mappingImageCapabilitySchemaDescriptorSource = map[string]ImageCapabilitySchemaDescriptorSourceEnum{
	"GLOBAL": ImageCapabilitySchemaDescriptorSourceGlobal,
	"IMAGE":  ImageCapabilitySchemaDescriptorSourceImage,
}

// GetImageCapabilitySchemaDescriptorSourceEnumValues Enumerates the set of values for ImageCapabilitySchemaDescriptorSourceEnum
func GetImageCapabilitySchemaDescriptorSourceEnumValues() []ImageCapabilitySchemaDescriptorSourceEnum {
	values := make([]ImageCapabilitySchemaDescriptorSourceEnum, 0)
	for _, v := range mappingImageCapabilitySchemaDescriptorSource {
		values = append(values, v)
	}
	return values
}
