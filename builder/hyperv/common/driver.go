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

	// Finds the IP address of a host adapter connected to switch
	GetHostAdapterIpAddressForSwitch(string) (string, error)

	// Type scan codes to virtual keyboard of vm
	TypeScanCodes(string, string) error
}
