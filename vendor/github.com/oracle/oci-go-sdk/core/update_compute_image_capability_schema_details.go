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

// UpdateComputeImageCapabilitySchemaDetails Create Image Capability Schema for an image.
type UpdateComputeImageCapabilitySchemaDetails struct {

	// A user-friendly name for the compute image capability schema
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The map of each capability name to its ImageCapabilitySchemaDescriptor.
	SchemaData map[string]ImageCapabilitySchemaDescriptor `mandatory:"false" json:"schemaData"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
}

func (m UpdateComputeImageCapabilitySchemaDetails) String() string {
	return common.PointerString(m)
}

// UnmarshalJSON unmarshals from json
func (m *UpdateComputeImageCapabilitySchemaDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		DisplayName  *string                                    `json:"displayName"`
		FreeformTags map[string]string                          `json:"freeformTags"`
		SchemaData   map[string]imagecapabilityschemadescriptor `json:"schemaData"`
		DefinedTags  map[string]map[string]interface{}          `json:"definedTags"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.DisplayName = model.DisplayName

	m.FreeformTags = model.FreeformTags

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

	m.DefinedTags = model.DefinedTags

	return
}
