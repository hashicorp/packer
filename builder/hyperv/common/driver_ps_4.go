package common

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"

	"github.com/hashicorp/packer/builder/hyperv/common/powershell"
	"github.com/hashicorp/packer/builder/hyperv/common/powershell/hyperv"
)

type HypervPS4Driver struct {
}

func NewHypervPS4Driver() (Driver, error) {
	appliesTo := "Applies to Windows 8.1, Windows PowerShell 4.0, Windows Server 2012 R2 only"

	// Check this is Windows
	if runtime.GOOS != "windows" {
		err := fmt.Errorf("%s", appliesTo)
		return nil, err
	}

	ps4Driver := &HypervPS4Driver{}

	if err := ps4Driver.Verify(); err != nil {
		return nil, err
	}

	return ps4Driver, nil
}

func (d *HypervPS4Driver) IsRunning(vmName string) (bool, error) {
	return hyperv.IsRunning(vmName)
}

func (d *HypervPS4Driver) IsOff(vmName string) (bool, error) {
	return hyperv.IsOff(vmName)
}

func (d *HypervPS4Driver) Uptime(vmName string) (uint64, error) {
	return hyperv.Uptime(vmName)
}

// Start starts a VM specified by the name given.
func (d *HypervPS4Driver) Start(vmName string) error {
	return hyperv.StartVirtualMachine(vmName)
}

// Stop stops a VM specified by the name given.
func (d *HypervPS4Driver) Stop(vmName string) error {
	return hyperv.StopVirtualMachine(vmName)
}

func (d *HypervPS4Driver) Verify() error {

	if err := d.verifyPSVersion(); err != nil {
		return err
	}

	if err := d.verifyPSHypervModule(); err != nil {
		return err
	}

	if err := d.verifyHypervPermissions(); err != nil {
		return err
	}

	return nil
}

// Get mac address for VM.
func (d *HypervPS4Driver) Mac(vmName string) (string, error) {
	res, err := hyperv.Mac(vmName)

	if err != nil {
		return res, err
	}

	if res == "" {
		err := fmt.Errorf("%s", "No mac address.")
		return res, err
	}

	return res, err
}

// Get ip address for mac address.
func (d *HypervPS4Driver) IpAddress(mac string) (string, error) {
	res, err := hyperv.IpAddress(mac)

	if err != nil {
		return res, err
	}

	if res == "" {
		err := fmt.Errorf("%s", "No ip address.")
		return res, err
	}
	return res, err
}

// Get host name from ip address
func (d *HypervPS4Driver) GetHostName(ip string) (string, error) {
	return powershell.GetHostName(ip)
}

func (d *HypervPS4Driver) GetVirtualMachineGeneration(vmName string) (uint, error) {
	return hyperv.GetVirtualMachineGeneration(vmName)
}

// Finds the IP address of a host adapter connected to switch
func (d *HypervPS4Driver) GetHostAdapterIpAddressForSwitch(switchName string) (string, error) {
	res, err := hyperv.GetHostAdapterIpAddressForSwitch(switchName)

	if err != nil {
		return res, err
	}

	if res == "" {
		err := fmt.Errorf("%s", "No ip address.")
		return res, err
	}
	return res, err
}

// Type scan codes to virtual keyboard of vm
func (d *HypervPS4Driver) TypeScanCodes(vmName string, scanCodes string) error {
	return hyperv.TypeScanCodes(vmName, scanCodes)
}

// Get network adapter address
func (d *HypervPS4Driver) GetVirtualMachineNetworkAdapterAddress(vmName string) (string, error) {
	return hyperv.GetVirtualMachineNetworkAdapterAddress(vmName)
}

//Set the vlan to use for switch
func (d *HypervPS4Driver) SetNetworkAdapterVlanId(switchName string, vlanId string) error {
	return hyperv.SetNetworkAdapterVlanId(switchName, vlanId)
}

//Set the vlan to use for machine
func (d *HypervPS4Driver) SetVirtualMachineVlanId(vmName string, vlanId string) error {
	return hyperv.SetVirtualMachineVlanId(vmName, vlanId)
}

func (d *HypervPS4Driver) SetVmNetworkAdapterMacAddress(vmName string, mac string) error {
	return hyperv.SetVmNetworkAdapterMacAddress(vmName, mac)
}

//Replace the network adapter with a (non-)legacy adapter
func (d *HypervPS4Driver) ReplaceVirtualMachineNetworkAdapter(vmName string, virtual bool) error {
	return hyperv.ReplaceVirtualMachineNetworkAdapter(vmName, virtual)
}

func (d *HypervPS4Driver) UntagVirtualMachineNetworkAdapterVlan(vmName string, switchName string) error {
	return hyperv.UntagVirtualMachineNetworkAdapterVlan(vmName, switchName)
}

func (d *HypervPS4Driver) CreateExternalVirtualSwitch(vmName string, switchName string) error {
	return hyperv.CreateExternalVirtualSwitch(vmName, switchName)
}

func (d *HypervPS4Driver) GetVirtualMachineSwitchName(vmName string) (string, error) {
	return hyperv.GetVirtualMachineSwitchName(vmName)
}

func (d *HypervPS4Driver) ConnectVirtualMachineNetworkAdapterToSwitch(vmName string, switchName string) error {
	return hyperv.ConnectVirtualMachineNetworkAdapterToSwitch(vmName, switchName)
}

func (d *HypervPS4Driver) DeleteVirtualSwitch(switchName string) error {
	return hyperv.DeleteVirtualSwitch(switchName)
}

func (d *HypervPS4Driver) CreateVirtualSwitch(switchName string, switchType string) (bool, error) {
	return hyperv.CreateVirtualSwitch(switchName, switchType)
}

func (d *HypervPS4Driver) AddVirtualMachineHardDrive(vmName string, vhdFile string, vhdName string,
	vhdSizeBytes int64, diskBlockSize int64, controllerType string) error {
	return hyperv.AddVirtualMachineHardDiskDrive(vmName, vhdFile, vhdName, vhdSizeBytes,
		diskBlockSize, controllerType)
}

func (d *HypervPS4Driver) CheckVMName(vmName string) error {
	return hyperv.CheckVMName(vmName)
}

func (d *HypervPS4Driver) CreateVirtualMachine(vmName string, path string, harddrivePath string, ram int64,
	diskSize int64, diskBlockSize int64, switchName string, generation uint, diffDisks bool,
	fixedVHD bool, version string) error {
	return hyperv.CreateVirtualMachine(vmName, path, harddrivePath, ram, diskSize, diskBlockSize, switchName,
		generation, diffDisks, fixedVHD, version)
}

func (d *HypervPS4Driver) CloneVirtualMachine(cloneFromVmcxPath string, cloneFromVmName string,
	cloneFromSnapshotName string, cloneAllSnapshots bool, vmName string, path string, harddrivePath string,
	ram int64, switchName string, copyTF bool) error {
	return hyperv.CloneVirtualMachine(cloneFromVmcxPath, cloneFromVmName, cloneFromSnapshotName,
		cloneAllSnapshots, vmName, path, harddrivePath, ram, switchName, copyTF)
}

func (d *HypervPS4Driver) DeleteVirtualMachine(vmName string) error {
	return hyperv.DeleteVirtualMachine(vmName)
}

func (d *HypervPS4Driver) SetVirtualMachineCpuCount(vmName string, cpu uint) error {
	return hyperv.SetVirtualMachineCpuCount(vmName, cpu)
}

func (d *HypervPS4Driver) SetVirtualMachineMacSpoofing(vmName string, enable bool) error {
	return hyperv.SetVirtualMachineMacSpoofing(vmName, enable)
}

func (d *HypervPS4Driver) SetVirtualMachineDynamicMemory(vmName string, enable bool) error {
	return hyperv.SetVirtualMachineDynamicMemory(vmName, enable)
}

func (d *HypervPS4Driver) SetVirtualMachineSecureBoot(vmName string, enable bool, templateName string) error {
	return hyperv.SetVirtualMachineSecureBoot(vmName, enable, templateName)
}

func (d *HypervPS4Driver) SetVirtualMachineVirtualizationExtensions(vmName string, enable bool) error {
	return hyperv.SetVirtualMachineVirtualizationExtensions(vmName, enable)
}

func (d *HypervPS4Driver) EnableVirtualMachineIntegrationService(vmName string,
	integrationServiceName string) error {
	return hyperv.EnableVirtualMachineIntegrationService(vmName, integrationServiceName)
}

func (d *HypervPS4Driver) ExportVirtualMachine(vmName string, path string) error {
	return hyperv.ExportVirtualMachine(vmName, path)
}

func (d *HypervPS4Driver) PreserveLegacyExportBehaviour(srcPath string, dstPath string) error {
	return hyperv.PreserveLegacyExportBehaviour(srcPath, dstPath)
}

func (d *HypervPS4Driver) MoveCreatedVHDsToOutputDir(srcPath string, dstPath string) error {
	return hyperv.MoveCreatedVHDsToOutputDir(srcPath, dstPath)
}

func (d *HypervPS4Driver) CompactDisks(path string) (result string, err error) {
	return hyperv.CompactDisks(path)
}

func (d *HypervPS4Driver) RestartVirtualMachine(vmName string) error {
	return hyperv.RestartVirtualMachine(vmName)
}

func (d *HypervPS4Driver) CreateDvdDrive(vmName string, isoPath string, generation uint) (uint, uint, error) {
	return hyperv.CreateDvdDrive(vmName, isoPath, generation)
}

func (d *HypervPS4Driver) MountDvdDrive(vmName string, path string, controllerNumber uint,
	controllerLocation uint) error {
	return hyperv.MountDvdDrive(vmName, path, controllerNumber, controllerLocation)
}

func (d *HypervPS4Driver) SetBootDvdDrive(vmName string, controllerNumber uint, controllerLocation uint,
	generation uint) error {
	return hyperv.SetBootDvdDrive(vmName, controllerNumber, controllerLocation, generation)
}

func (d *HypervPS4Driver) SetFirstBootDevice(vmName string, controllerType string, controllerNumber uint,
	controllerLocation uint, generation uint) error {
	return hyperv.SetFirstBootDevice(vmName, controllerType, controllerNumber, controllerLocation, generation)
}

func (d *HypervPS4Driver) SetBootOrder(vmName string, bootOrder []string) error {
	return hyperv.SetBootOrder(vmName, bootOrder)
}

func (d *HypervPS4Driver) UnmountDvdDrive(vmName string, controllerNumber uint, controllerLocation uint) error {
	return hyperv.UnmountDvdDrive(vmName, controllerNumber, controllerLocation)
}

func (d *HypervPS4Driver) DeleteDvdDrive(vmName string, controllerNumber uint, controllerLocation uint) error {
	return hyperv.DeleteDvdDrive(vmName, controllerNumber, controllerLocation)
}

func (d *HypervPS4Driver) MountFloppyDrive(vmName string, path string) error {
	return hyperv.MountFloppyDrive(vmName, path)
}

func (d *HypervPS4Driver) UnmountFloppyDrive(vmName string) error {
	return hyperv.UnmountFloppyDrive(vmName)
}

func (d *HypervPS4Driver) verifyPSVersion() error {

	log.Printf("Enter method: %s", "verifyPSVersion")
	// check PS is available and is of proper version
	versionCmd := "$host.version.Major"

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(versionCmd)
	if err != nil {
		return err
	}

	versionOutput := strings.TrimSpace(cmdOut)
	log.Printf("%s output: %s", versionCmd, versionOutput)

	ver, err := strconv.ParseInt(versionOutput, 10, 32)

	if err != nil {
		return err
	}

	if ver < 4 {
		err := fmt.Errorf("%s", "Windows PowerShell version 4.0 or higher is expected")
		return err
	}

	return nil
}

func (d *HypervPS4Driver) verifyPSHypervModule() error {

	log.Printf("Enter method: %s", "verifyPSHypervModule")

	versionCmd := "function foo(){try{ $commands = Get-Command -Module Hyper-V;if($commands.Length -eq 0){return $false} }catch{return $false}; return $true} foo"

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(versionCmd)
	if err != nil {
		return err
	}

	if powershell.IsFalse(cmdOut) {
		err := fmt.Errorf("%s", "PS Hyper-V module is not loaded. Make sure Hyper-V feature is on.")
		return err
	}

	return nil
}

func (d *HypervPS4Driver) isCurrentUserAHyperVAdministrator() (bool, error) {
	//SID:S-1-5-32-578 = 'BUILTIN\Hyper-V Administrators'
	//https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems

	var script = `
$identity = [System.Security.Principal.WindowsIdentity]::GetCurrent()
$principal = new-object System.Security.Principal.WindowsPrincipal($identity)
$hypervrole = [System.Security.Principal.SecurityIdentifier]"S-1-5-32-578"
return $principal.IsInRole($hypervrole)
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script)
	if err != nil {
		return false, err
	}

	return powershell.IsTrue(cmdOut), nil
}

func (d *HypervPS4Driver) verifyHypervPermissions() error {

	log.Printf("Enter method: %s", "verifyHypervPermissions")

	hyperVAdmin, err := d.isCurrentUserAHyperVAdministrator()
	if err != nil {
		log.Printf("Error discovering if current is is a Hyper-V Admin: %s", err)
	}
	if !hyperVAdmin {

		isAdmin, _ := powershell.IsCurrentUserAnAdministrator()

		if !isAdmin {
			err := fmt.Errorf("%s", "Current user is not a member of 'Hyper-V Administrators' or 'Administrators' group")
			return err
		}
	}

	return nil
}

// Connect connects to a VM specified by the name given.
func (d *HypervPS4Driver) Connect(vmName string) (context.CancelFunc, error) {
	return hyperv.ConnectVirtualMachine(vmName)
}

// Disconnect disconnects to a VM specified by calling the context cancel function returned
// from Connect.
func (d *HypervPS4Driver) Disconnect(cancel context.CancelFunc) {
	hyperv.DisconnectVirtualMachine(cancel)
}
