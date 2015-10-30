// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

// A driver is able to talk to HyperV and perform certain
// operations with it. Some of the operations on here may seem overly
// specific, but they were built specifically in mind to handle features
// of the HyperV builder for Packer, and to abstract differences in
// versions out of the builder steps, so sometimes the methods are
// extremely specific.
type Driver interface {

	// Checks if the VM named is running.
	IsRunning(string) (bool, error)

	// Checks if the VM named is off.
	IsOff(string) (bool, error)

	//How long has VM been on
	Uptime(vmName string) (uint64, error)

	// Start starts a VM specified by the name given.
	Start(string) error

	// Stop stops a VM specified by the name given.
	Stop(string) error

	// Verify checks to make sure that this driver should function
	// properly. If there is any indication the driver can't function,
	// this will return an error.
	Verify() error

	// Finds the MAC address of the NIC nic0
	Mac(string) (string, error)

	// Finds the IP address of a VM connected that uses DHCP by its MAC address
	IpAddress(string) (string, error)

	// Finds the hostname for the ip address
	GetHostName(string) (string, error)

	// Finds the IP address of a host adapter connected to switch
	GetHostAdapterIpAddressForSwitch(string) (string, error)

	// Type scan codes to virtual keyboard of vm
	TypeScanCodes(string, string) error

	//Get the ip address for network adaptor
	GetVirtualMachineNetworkAdapterAddress(string) (string, error)

	//Set the vlan to use for switch
	SetNetworkAdapterVlanId(string, string) error

	//Set the vlan to use for machine
	SetVirtualMachineVlanId(string, string) error

	UntagVirtualMachineNetworkAdapterVlan(string, string) error

	CreateExternalVirtualSwitch(string, string) error

	GetVirtualMachineSwitchName(string) (string, error)

	ConnectVirtualMachineNetworkAdapterToSwitch(string, string) error

	CreateVirtualSwitch(string, string) (bool, error)

	DeleteVirtualSwitch(string) error

	CreateVirtualMachine(string, string, int64, int64, string, uint) error

	DeleteVirtualMachine(string) error

	SetVirtualMachineCpu(string, uint) error

	SetSecureBoot(string, bool) error

	EnableVirtualMachineIntegrationService(string, string) error

	ExportVirtualMachine(string, string) error

	CompactDisks(string, string) error

	CopyExportedVirtualMachine(string, string, string, string) error

	RestartVirtualMachine(string) error

	CreateDvdDrive(string, uint) (uint, uint, error)

	MountDvdDrive(string, string) error

	MountDvdDriveByLocation(string, string, uint, uint) error

	UnmountDvdDrive(string) error

	DeleteDvdDrive(string, string, string) error

	UnmountFloppyDrive(vmName string) error
}
