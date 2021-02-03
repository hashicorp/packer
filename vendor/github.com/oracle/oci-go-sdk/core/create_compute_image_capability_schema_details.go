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

// CreateComputeImageCapabilitySchemaDetails Create Image Capability Schema for an image.
type CreateComputeImageCapabilitySchemaDetails struct {

	// The OCID of the compartment that contains the resource.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The name of the compute global image capability schema version
	ComputeGlobalImageCapabilitySchemaVersionName *string `mandatory:"true" json:"computeGlobalImageCapabilitySchemaVersionName"`

	// The ocid of the image
	ImageId *string `mandatory:"true" json:"imageId"`

	// The map of each capability name to its ImageCapabilitySchemaDescriptor.
	SchemaData map[string]ImageCapabilitySchemaDescriptor `mandatory:"true" json:"schemaData"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// A user-friendly name for the compute image capability schema
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m CreateComputeImageCapabilitySchemaDetails) String() string {
	return common.PointerString(m)
}

// UnmarshalJSON unmarshals from json
func (m *CreateComputeImageCapabilitySchemaDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		FreeformTags                                  map[string]string                          `json:"freeformTags"`
		DisplayName                                   *string                                    `json:"displayName"`
		DefinedTags                                   map[string]map[string]interface{}          `json:"definedTags"`
		CompartmentId                                 *string                                    `json:"compartmentId"`
		ComputeGlobalImageCapabilitySchemaVersionName *string                                    `json:"computeGlobalImageCapabilitySchemaVersionName"`
		ImageId                                       *string                                    `json:"imageId"`
		SchemaData                                    map[string]imagecapabilityschemadescriptor `json:"schemaData"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.FreeformTags = model.FreeformTags

	m.DisplayName = model.DisplayName

	m.DefinedTags = model.DefinedTags

	m.CompartmentId = model.CompartmentId

	m.ComputeGlobalImageCapabilitySchemaVersionName = model.ComputeGlobalImageCapabilitySchemaVersionName

	m.ImageId = model.ImageId

	m.SchemaData = make(map[string]ImageCapabilitySchemaDescriptor)
	for k, v := range model.SchemaData {
		nn, e = v.UnmarshalPolymorphicJSON(v.JsonData)
		if e != nil {
			return e
		}
		if nn != nil {
			m.SchemaData[k] = nn.(ImageCapabilitySchemaDescriptor)
		} else {
			m.SchemaData[k] = nil
		}
	}

	return
}
