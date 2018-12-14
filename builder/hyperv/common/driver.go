package common

import (
	"context"
)

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

	SetVmNetworkAdapterMacAddress(string, string) error

	UntagVirtualMachineNetworkAdapterVlan(string, string) error

	CreateExternalVirtualSwitch(string, string) error

	GetVirtualMachineSwitchName(string) (string, error)

	ConnectVirtualMachineNetworkAdapterToSwitch(string, string) error

	CreateVirtualSwitch(string, string) (bool, error)

	DeleteVirtualSwitch(string) error

	CreateVirtualMachine(string, string, string, int64, int64, int64, string, uint, bool, bool) error

	AddVirtualMachineHardDrive(string, string, string, int64, int64, string) error

	CloneVirtualMachine(string, string, string, bool, string, string, string, int64, string, bool) error

	DeleteVirtualMachine(string) error

	GetVirtualMachineGeneration(string) (uint, error)

	SetVirtualMachineCpuCount(string, uint) error

	SetVirtualMachineMacSpoofing(string, bool) error

	SetVirtualMachineDynamicMemory(string, bool) error

	SetVirtualMachineSecureBoot(string, bool, string) error

	SetVirtualMachineVirtualizationExtensions(string, bool) error

	EnableVirtualMachineIntegrationService(string, string) error

	ExportVirtualMachine(string, string) error

	PreserveLegacyExportBehaviour(string, string) error

	MoveCreatedVHDsToOutputDir(string, string) error

	CompactDisks(string) (string, error)

	RestartVirtualMachine(string) error

	CreateDvdDrive(string, string, uint) (uint, uint, error)

	MountDvdDrive(string, string, uint, uint) error

	SetBootDvdDrive(string, uint, uint, uint) error

	UnmountDvdDrive(string, uint, uint) error

	DeleteDvdDrive(string, uint, uint) error

	MountFloppyDrive(string, string) error

	UnmountFloppyDrive(string) error

	// Connect connects to a VM specified by the name given.
	Connect(string) (context.CancelFunc, error)

	// Disconnect disconnects to a VM specified by the context cancel function.
	Disconnect(context.CancelFunc)
}
