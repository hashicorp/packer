// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"github.com/oracle/oci-go-sdk/common"
)

// Image A boot disk image for launching an instance. For more information, see
// Overview of the Compute Service (https://docs.us-phoenix-1.oraclecloud.com/Content/Compute/Concepts/computeoverview.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access, see
// Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type Image struct {

	// The OCID of the compartment containing the instance you want to use as the basis for the image.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// Whether instances launched with this image can be used to create new images.
	// For example, you cannot create an image of an Oracle Database instance.
	// Example: `true`
	CreateImageAllowed *bool `mandatory:"true" json:"createImageAllowed"`

	// The OCID of the image.
	Id *string `mandatory:"true" json:"id"`

	LifecycleState ImageLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The image's operating system.
	// Example: `Oracle Linux`
	OperatingSystem *string `mandatory:"true" json:"operatingSystem"`

	// The image's operating system version.
	// Example: `7.2`
	OperatingSystemVersion *string `mandatory:"true" json:"operatingSystemVersion"`

	// The date and time the image was created, in the format defined by RFC3339.
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`

	// The OCID of the image originally used to launch the instance.
	BaseImageId *string `mandatory:"false" json:"baseImageId"`

	// A user-friendly name for the image. It does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	// You cannot use an Oracle-provided image name as a custom image name.
	// Example: `My custom Oracle Linux image`
	DisplayName *string `mandatory:"false" json:"displayName"`
}

func (m Image) String() string {
	return common.PointerString(m)
}

// ImageLifecycleStateEnum Enum with underlying type: string
type ImageLifecycleStateEnum string

// Set of constants representing the allowable values for ImageLifecycleState
const (
	ImageLifecycleStateProvisioning ImageLifecycleStateEnum = "PROVISIONING"
	ImageLifecycleStateImporting    ImageLifecycleStateEnum = "IMPORTING"
	ImageLifecycleStateAvailable    ImageLifecycleStateEnum = "AVAILABLE"
	ImageLifecycleStateExporting    ImageLifecycleStateEnum = "EXPORTING"
	ImageLifecycleStateDisabled     ImageLifecycleStateEnum = "DISABLED"
	ImageLifecycleStateDeleted      ImageLifecycleStateEnum = "DELETED"
)

var mappingImageLifecycleState = map[string]ImageLifecycleStateEnum{
	"PROVISIONING": ImageLifecycleStateProvisioning,
	"IMPORTING":    ImageLifecycleStateImporting,
	"AVAILABLE":    ImageLifecycleStateAvailable,
	"EXPORTING":    ImageLifecycleStateExporting,
	"DISABLED":     ImageLifecycleStateDisabled,
	"DELETED":      ImageLifecycleStateDeleted,
}

// GetImageLifecycleStateEnumValues Enumerates the set of values for ImageLifecycleState
func GetImageLifecycleStateEnumValues() []ImageLifecycleStateEnum {
	values := make([]ImageLifecycleStateEnum, 0)
	for _, v := range mappingImageLifecycleState {
		values = append(values, v)
	}
	return values
}
