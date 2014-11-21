// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package model

import "encoding/xml"

type Deployment struct {
	XMLName   				xml.Name 			`xml:"Deployment"`
	Xmlns	  				string 				`xml:"xmlns,attr"`
	Name 					string
	DeploymentSlot 			string
	PrivateID 				string
	Status 					string
	Label 					string
	Url 					string
	Configuration 			string
	RoleInstanceList 		[]RoleInstance		`xml:"RoleInstanceList>RoleInstance"`
	UpgradeDomainCount		string
	RoleList 				[]Role				`xml:"RoleList>Role"`
	SdkVersion				string
	Locked					string
	RollbackAllowed			string
	CreatedTime				string
	LastModifiedTime		string
	ExtendedProperties		[]ExtendedProperty	`xml:"ExtendedProperties"`
	PersistentVMDowntime	PersistentVMDowntime
	VirtualIPs				[]VirtualIP			`xml:"VirtualIPs>VirtualIP"`
	ExtensionConfiguration 	ExtensionConfiguration
}

type PersistentVMDowntime struct {
	StartTime string
	EndTime string
	Status string
}

type ExtensionConfiguration struct {
	AllRoles	Extension	`xml:"AllRoles>Extension"`
	NamedRoles	[]Role1 	`xml:"NamedRoles>Role"`
}

type Role1 struct {
	RoleName	string
	Extensions	[]Extension	`xml:"Extensions>Extension"`
}

type Extension struct {
	Id	string
	SequenceNumber	string
	State	string
}

type RoleInstance struct {
	RoleName  							string
	InstanceName						string
	InstanceStatus						string
	ExtendedInstanceStatus				string
	InstanceUpgradeDomain				string
	InstanceFaultDomain					string
	InstanceSize						string
	InstanceStateDetails				string
	IpAddress							string
	InstanceEndpoints					[]InstanceEndpoint	`xml:"InstanceEndpoints>InstanceEndpoint"`
	PowerState							string
	HostName							string
	RemoteAccessCertificateThumbprint	string
	GuestAgentStatus					GuestAgentStatus
	ResourceExtensionStatusList			[]ResourceExtensionStatus `xml:"ResourceExtensionStatusList>ResourceExtensionStatus"`
}

type ResourceExtensionStatus struct {
	HandlerName  			string
	Version  				string
	Status  				string
	Code  					string
	FormattedMessage  		FormattedMessage
	ExtensionSettingStatus  ExtensionSettingStatus

}

type ExtensionSettingStatus struct {
	Timestamp  			string
	Name  				string
	Operation  			string
	Status  			string
	Code  				string
	FormattedMessage  	FormattedMessage
	SubStatusList  		[]SubStatus  `xml:"SubStatusList>SubStatus"`

}

type SubStatus struct {
	Name  				string
	Status  			string
	FormattedMessage  	FormattedMessage
}

type GuestAgentStatus struct {
	ProtocolVersion  	string
	Timestamp  			string
	GuestAgentVersion  	string
	Status  			string
	FormattedMessage  	FormattedMessage
}

type FormattedMessage struct {
	Language  			string
	Message  			string
}

type InstanceEndpoint struct {
	Name  			string
	Vip  			string
	PublicPort  	string
	LocalPort  		string
	Protocol  		string
}

type Role struct {
	RoleName				string
	OsVersion				string
	RoleType				string
	ConfigurationSets		[]ConfigurationSet		`xml:"ConfigurationSets>ConfigurationSet"`
	VMImageName				string
	DataVirtualHardDisks	[]DataVirtualHardDisk	`xml:"DataVirtualHardDisks>DataVirtualHardDisk"`
	OSVirtualHardDisk		OSVirtualHardDisk
	RoleSize				string
	ProvisionGuestAgent		bool
}

type ConfigurationSet struct {
//	Type					string 	`xml:"i:type,attr"`
	ConfigurationSetType	string
	InputEndpoints			[]InputEndpoint	`xml:"InputEndpoints>InputEndpoint"`
}

type InputEndpoint struct {
	LocalPort	string
	Name		string
	Port		string
	Protocol	string
	Vip			string
}

type DataVirtualHardDisk struct {
	HostCaching	string
	DiskName	string
	Lun	string
	LogicalDiskSizeInGB	string
	MediaLink	string
}

type OSVirtualHardDisk struct {
	HostCaching	string
	DiskName	string
	MediaLink	string
	SourceImageName	string
	OS	string
}

type ExtendedProperty struct {
	Name	string
	Value	string
}

type VirtualIP struct {
	DataVirtualHardDisks	string
	IsDnsProgrammed			string
	Name					string
}
