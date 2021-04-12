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

// ExportImageDetails The destination details for the image export.
// Set `destinationType` to `objectStorageTuple`
// and use ExportImageViaObjectStorageTupleDetails
// when specifying the namespace, bucket name, and object name.
// Set `destinationType` to `objectStorageUri` and
// use ExportImageViaObjectStorageUriDetails
// when specifying the Object Storage URL.
type ExportImageDetails interface {

	// The format of the image to be exported. The default value is "OCI".
	GetExportFormat() ExportImageDetailsExportFormatEnum
}

type exportimagedetails struct {
	JsonData        []byte
	ExportFormat    ExportImageDetailsExportFormatEnum `mandatory:"false" json:"exportFormat,omitempty"`
	DestinationType string                             `json:"destinationType"`
}

// UnmarshalJSON unmarshals json
func (m *exportimagedetails) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerexportimagedetails exportimagedetails
	s := struct {
		Model Unmarshalerexportimagedetails
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.ExportFormat = s.Model.ExportFormat
	m.DestinationType = s.Model.DestinationType

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *exportimagedetails) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.DestinationType {
	case "objectStorageUri":
		mm := ExportImageViaObjectStorageUriDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "objectStorageTuple":
		mm := ExportImageViaObjectStorageTupleDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

//GetExportFormat returns ExportFormat
func (m exportimagedetails) GetExportFormat() ExportImageDetailsExportFormatEnum {
	return m.ExportFormat
}

func (m exportimagedetails) String() string {
	return common.PointerString(m)
}

// ExportImageDetailsExportFormatEnum Enum with underlying type: string
type ExportImageDetailsExportFormatEnum string

// Set of constants representing the allowable values for ExportImageDetailsExportFormatEnum
const (
	ExportImageDetailsExportFormatQcow2 ExportImageDetailsExportFormatEnum = "QCOW2"
	ExportImageDetailsExportFormatVmdk  ExportImageDetailsExportFormatEnum = "VMDK"
	ExportImageDetailsExportFormatOci   ExportImageDetailsExportFormatEnum = "OCI"
	ExportImageDetailsExportFormatVhd   ExportImageDetailsExportFormatEnum = "VHD"
	ExportImageDetailsExportFormatVdi   ExportImageDetailsExportFormatEnum = "VDI"
)

var mappingExportImageDetailsExportFormat = map[string]ExportImageDetailsExportFormatEnum{
	"QCOW2": ExportImageDetailsExportFormatQcow2,
	"VMDK":  ExportImageDetailsExportFormatVmdk,
	"OCI":   ExportImageDetailsExportFormatOci,
	"VHD":   ExportImageDetailsExportFormatVhd,
	"VDI":   ExportImageDetailsExportFormatVdi,
}

// GetExportImageDetailsExportFormatEnumValues Enumerates the set of values for ExportImageDetailsExportFormatEnum
func GetExportImageDetailsExportFormatEnumValues() []ExportImageDetailsExportFormatEnum {
	values := make([]ExportImageDetailsExportFormatEnum, 0)
	for _, v := range mappingExportImageDetailsExportFormat {
		values = append(values, v)
	}
	return values
}
