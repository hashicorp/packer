// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// File Storage Service API
//
// The API for the File Storage Service.
//

package filestorage

import (
	"github.com/oracle/oci-go-sdk/common"
)

// CreateMountTargetDetails The representation of CreateMountTargetDetails
type CreateMountTargetDetails struct {

	// The availability domain in which to create the mount target.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment in which to create the mount target.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the subnet in which to create the mount target.
	SubnetId *string `mandatory:"true" json:"subnetId"`

	// A user-friendly name. It does not have to be unique, and it is changeable.
	// Avoid entering confidential information.
	// Example: `My mount target`
	DisplayName *string `mandatory:"false" json:"displayName"`

	// The hostname for the mount target's IP address, used for
	// DNS resolution. The value is the hostname portion of the private IP
	// address's fully qualified domain name (FQDN). For example,
	// `files-1` in the FQDN `files-1.subnet123.vcn1.oraclevcn.com`.
	// Must be unique across all VNICs in the subnet and comply
	// with RFC 952 (https://tools.ietf.org/html/rfc952)
	// and RFC 1123 (https://tools.ietf.org/html/rfc1123).
	// For more information, see
	// DNS in Your Virtual Cloud Network (https://docs.us-phoenix-1.oraclecloud.com/Content/Network/Concepts/dns.htm).
	// Example: `files-1`
	HostnameLabel *string `mandatory:"false" json:"hostnameLabel"`

	// A private IP address of your choice. Must be an available IP address within
	// the subnet's CIDR. If you don't specify a value, Oracle automatically
	// assigns a private IP address from the subnet.
	// Example: `10.0.3.3`
	IpAddress *string `mandatory:"false" json:"ipAddress"`
}

func (m CreateMountTargetDetails) String() string {
	return common.PointerString(m)
}
