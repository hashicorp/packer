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

// LaunchInstanceDetails Instance launch details.
// Use the `sourceDetails` parameter to specify whether a boot volume or an image should be used to launch a new instance.
type LaunchInstanceDetails struct {

	// The availability domain of the instance.
	// Example: `Uocm:PHX-AD-1`
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The shape of an instance. The shape determines the number of CPUs, amount of memory,
	// and other resources allocated to the instance.
	// You can enumerate all available shapes by calling ListShapes.
	Shape *string `mandatory:"true" json:"shape"`

	// Details for the primary VNIC, which is automatically created and attached when
	// the instance is launched.
	CreateVnicDetails *CreateVnicDetails `mandatory:"false" json:"createVnicDetails"`

	// The OCID of the dedicated VM host.
	DedicatedVmHostId *string `mandatory:"false" json:"dedicatedVmHostId"`

	// Defined tags for this resource. Each key is predefined and scoped to a
	// namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// A user-friendly name. Does not have to be unique, and it's changeable.
	// Avoid entering confidential information.
	// Example: `My bare metal instance`
	DisplayName *string `mandatory:"false" json:"displayName"`

	// Additional metadata key/value pairs that you provide. They serve the same purpose and
	// functionality as fields in the `metadata` object.
	// They are distinguished from `metadata` fields in that these can be nested JSON objects
	// (whereas `metadata` fields are string/string maps only).
	// The combined size of the `metadata` and `extendedMetadata` objects can be a maximum of
	// 32,000 bytes.
	ExtendedMetadata map[string]interface{} `mandatory:"false" json:"extendedMetadata"`

	// A fault domain is a grouping of hardware and infrastructure within an availability domain.
	// Each availability domain contains three fault domains. Fault domains let you distribute your
	// instances so that they are not on the same physical hardware within a single availability domain.
	// A hardware failure or Compute hardware maintenance that affects one fault domain does not affect
	// instances in other fault domains.
	// If you do not specify the fault domain, the system selects one for you.
	//
	// To get a list of fault domains, use the
	// ListFaultDomains operation in the
	// Identity and Access Management Service API.
	// Example: `FAULT-DOMAIN-1`
	FaultDomain *string `mandatory:"false" json:"faultDomain"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no
	// predefined name, type, or namespace. For more information, see Resource Tags (https://docs.cloud.oracle.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// Deprecated. Instead use `hostnameLabel` in
	// CreateVnicDetails.
	// If you provide both, the values must match.
	HostnameLabel *string `mandatory:"false" json:"hostnameLabel"`

	// Deprecated. Use `sourceDetails` with InstanceSourceViaImageDetails
	// source type instead. If you specify values for both, the values must match.
	ImageId *string `mandatory:"false" json:"imageId"`

	// This is an advanced option.
	// When a bare metal or virtual machine
	// instance boots, the iPXE firmware that runs on the instance is
	// configured to run an iPXE script to continue the boot process.
	// If you want more control over the boot process, you can provide
	// your own custom iPXE script that will run when the instance boots;
	// however, you should be aware that the same iPXE script will run
	// every time an instance boots; not only after the initial
	// LaunchInstance call.
	// The default iPXE script connects to the instance's local boot
	// volume over iSCSI and performs a network boot. If you use a custom iPXE
	// script and want to network-boot from the instance's local boot volume
	// over iSCSI the same way as the default iPXE script, you should use the
	// following iSCSI IP address: 169.254.0.2, and boot volume IQN:
	// iqn.2015-02.oracle.boot.
	// For more information about the Bring Your Own Image feature of
	// Oracle Cloud Infrastructure, see
	// Bring Your Own Image (https://docs.cloud.oracle.com/Content/Compute/References/bringyourownimage.htm).
	// For more information about iPXE, see http://ipxe.org.
	IpxeScript *string `mandatory:"false" json:"ipxeScript"`

	// Options for tuning the compatibility and performance of VM shapes. The values that you specify override any
	// default values.
	LaunchOptions *LaunchOptions `mandatory:"false" json:"launchOptions"`

	AvailabilityConfig *LaunchInstanceAvailabilityConfigDetails `mandatory:"false" json:"availabilityConfig"`

	// Custom metadata key/value pairs that you provide, such as the SSH public key
	// required to connect to the instance.
	// A metadata service runs on every launched instance. The service is an HTTP
	// endpoint listening on 169.254.169.254. You can use the service to:
	// * Provide information to Cloud-Init (https://cloudinit.readthedocs.org/en/latest/)
	//   to be used for various system initialization tasks.
	// * Get information about the instance, including the custom metadata that you
	//   provide when you launch the instance.
	//  **Providing Cloud-Init Metadata**
	//  You can use the following metadata key names to provide information to
	//  Cloud-Init:
	//  **"ssh_authorized_keys"** - Provide one or more public SSH keys to be
	//  included in the `~/.ssh/authorized_keys` file for the default user on the
	//  instance. Use a newline character to separate multiple keys. The SSH
	//  keys must be in the format necessary for the `authorized_keys` file, as shown
	//  in the example below.
	//  **"user_data"** - Provide your own base64-encoded data to be used by
	//  Cloud-Init to run custom scripts or provide custom Cloud-Init configuration. For
	//  information about how to take advantage of user data, see the
	//  Cloud-Init Documentation (http://cloudinit.readthedocs.org/en/latest/topics/format.html).
	//  **Metadata Example**
	//       "metadata" : {
	//          "quake_bot_level" : "Severe",
	//          "ssh_authorized_keys" : "ssh-rsa <your_public_SSH_key>== rsa-key-20160227",
	//          "user_data" : "<your_public_SSH_key>=="
	//       }
	//  **Getting Metadata on the Instance**
	//  To get information about your instance, connect to the instance using SSH and issue any of the
	//  following GET requests:
	//      curl -H "Authorization: Bearer Oracle" http://169.254.169.254/opc/v2/instance/
	//      curl -H "Authorization: Bearer Oracle" http://169.254.169.254/opc/v2/instance/metadata/
	//      curl -H "Authorization: Bearer Oracle" http://169.254.169.254/opc/v2/instance/metadata/<any-key-name>
	//  You'll get back a response that includes all the instance information; only the metadata information; or
	//  the metadata information for the specified key name, respectively.
	//  The combined size of the `metadata` and `extendedMetadata` objects can be a maximum of 32,000 bytes.
	Metadata map[string]string `mandatory:"false" json:"metadata"`

	AgentConfig *LaunchInstanceAgentConfigDetails `mandatory:"false" json:"agentConfig"`

	ShapeConfig *LaunchInstanceShapeConfigDetails `mandatory:"false" json:"shapeConfig"`

	// Details for creating an instance.
	// Use this parameter to specify whether a boot volume or an image should be used to launch a new instance.
	SourceDetails InstanceSourceDetails `mandatory:"false" json:"sourceDetails"`

	// Deprecated. Instead use `subnetId` in
	// CreateVnicDetails.
	// At least one of them is required; if you provide both, the values must match.
	SubnetId *string `mandatory:"false" json:"subnetId"`

	// Whether to enable in-transit encryption for the data volume's paravirtualized attachment. The default value is false.
	IsPvEncryptionInTransitEnabled *bool `mandatory:"false" json:"isPvEncryptionInTransitEnabled"`
}

func (m LaunchInstanceDetails) String() string {
	return common.PointerString(m)
}

// UnmarshalJSON unmarshals from json
func (m *LaunchInstanceDetails) UnmarshalJSON(data []byte) (e error) {
	model := struct {
		CreateVnicDetails              *CreateVnicDetails                       `json:"createVnicDetails"`
		DedicatedVmHostId              *string                                  `json:"dedicatedVmHostId"`
		DefinedTags                    map[string]map[string]interface{}        `json:"definedTags"`
		DisplayName                    *string                                  `json:"displayName"`
		ExtendedMetadata               map[string]interface{}                   `json:"extendedMetadata"`
		FaultDomain                    *string                                  `json:"faultDomain"`
		FreeformTags                   map[string]string                        `json:"freeformTags"`
		HostnameLabel                  *string                                  `json:"hostnameLabel"`
		ImageId                        *string                                  `json:"imageId"`
		IpxeScript                     *string                                  `json:"ipxeScript"`
		LaunchOptions                  *LaunchOptions                           `json:"launchOptions"`
		AvailabilityConfig             *LaunchInstanceAvailabilityConfigDetails `json:"availabilityConfig"`
		Metadata                       map[string]string                        `json:"metadata"`
		AgentConfig                    *LaunchInstanceAgentConfigDetails        `json:"agentConfig"`
		ShapeConfig                    *LaunchInstanceShapeConfigDetails        `json:"shapeConfig"`
		SourceDetails                  instancesourcedetails                    `json:"sourceDetails"`
		SubnetId                       *string                                  `json:"subnetId"`
		IsPvEncryptionInTransitEnabled *bool                                    `json:"isPvEncryptionInTransitEnabled"`
		AvailabilityDomain             *string                                  `json:"availabilityDomain"`
		CompartmentId                  *string                                  `json:"compartmentId"`
		Shape                          *string                                  `json:"shape"`
	}{}

	e = json.Unmarshal(data, &model)
	if e != nil {
		return
	}
	var nn interface{}
	m.CreateVnicDetails = model.CreateVnicDetails

	m.DedicatedVmHostId = model.DedicatedVmHostId

	m.DefinedTags = model.DefinedTags

	m.DisplayName = model.DisplayName

	m.ExtendedMetadata = model.ExtendedMetadata

	m.FaultDomain = model.FaultDomain

	m.FreeformTags = model.FreeformTags

	m.HostnameLabel = model.HostnameLabel

	m.ImageId = model.ImageId

	m.IpxeScript = model.IpxeScript

	m.LaunchOptions = model.LaunchOptions

	m.AvailabilityConfig = model.AvailabilityConfig

	m.Metadata = model.Metadata

	m.AgentConfig = model.AgentConfig

	m.ShapeConfig = model.ShapeConfig

	nn, e = model.SourceDetails.UnmarshalPolymorphicJSON(model.SourceDetails.JsonData)
	if e != nil {
		return
	}
	if nn != nil {
		m.SourceDetails = nn.(InstanceSourceDetails)
	} else {
		m.SourceDetails = nil
	}

	m.SubnetId = model.SubnetId

	m.IsPvEncryptionInTransitEnabled = model.IsPvEncryptionInTransitEnabled

	m.AvailabilityDomain = model.AvailabilityDomain

	m.CompartmentId = model.CompartmentId

	m.Shape = model.Shape

	return
}
