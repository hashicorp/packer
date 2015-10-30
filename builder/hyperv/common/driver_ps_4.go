// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/packer/powershell"
	"github.com/mitchellh/packer/powershell/hyperv"
	"log"
	"runtime"
	"strconv"
	"strings"
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

	if err := d.verifyElevatedMode(); err != nil {
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

func (d *HypervPS4Driver) CreateVirtualMachine(vmName string, path string, ram int64, diskSize int64, switchName string, generation uint) error {
	return hyperv.CreateVirtualMachine(vmName, path, ram, diskSize, switchName, generation)
}

func (d *HypervPS4Driver) DeleteVirtualMachine(vmName string) error {
	return hyperv.DeleteVirtualMachine(vmName)
}

func (d *HypervPS4Driver) SetVirtualMachineCpu(vmName string, cpu uint) error {
	return hyperv.SetVirtualMachineCpu(vmName, cpu)
}

func (d *HypervPS4Driver) SetSecureBoot(vmName string, enable bool) error {
	return hyperv.SetSecureBoot(vmName, enable)
}

func (d *HypervPS4Driver) EnableVirtualMachineIntegrationService(vmName string, integrationServiceName string) error {
	return hyperv.EnableVirtualMachineIntegrationService(vmName, integrationServiceName)
}

func (d *HypervPS4Driver) ExportVirtualMachine(vmName string, path string) error {
	return hyperv.ExportVirtualMachine(vmName, path)
}

func (d *HypervPS4Driver) CompactDisks(expPath string, vhdDir string) error {
	return hyperv.CompactDisks(expPath, vhdDir)
}

func (d *HypervPS4Driver) CopyExportedVirtualMachine(expPath string, outputPath string, vhdDir string, vmDir string) error {
	return hyperv.CopyExportedVirtualMachine(expPath, outputPath, vhdDir, vmDir)
}

func (d *HypervPS4Driver) RestartVirtualMachine(vmName string) error {
	return hyperv.RestartVirtualMachine(vmName)
}

func (d *HypervPS4Driver) CreateDvdDrive(vmName string, generation uint) (uint, uint, error) {
	return hyperv.CreateDvdDrive(vmName, generation)
}

func (d *HypervPS4Driver) MountDvdDrive(vmName string, path string) error {
	return hyperv.MountDvdDrive(vmName, path)
}

func (d *HypervPS4Driver) MountDvdDriveByLocation(vmName string, path string, controllerNumber uint, controllerLocation uint) error {
	return hyperv.MountDvdDriveByLocation(vmName, path, controllerNumber, controllerLocation)
}

func (d *HypervPS4Driver) SetBootDvdDrive(vmName string, controllerNumber uint, controllerLocation uint) error {
	return hyperv.SetBootDvdDrive(vmName, controllerNumber, controllerLocation)
}

func (d *HypervPS4Driver) UnmountDvdDrive(vmName string) error {
	return hyperv.UnmountDvdDrive(vmName)
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

	versionOutput := strings.TrimSpace(string(cmdOut))
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

	res := strings.TrimSpace(string(cmdOut))

	if res == "False" {
		err := fmt.Errorf("%s", "PS Hyper-V module is not loaded. Make sure Hyper-V feature is on.")
		return err
	}

	return nil
}

func (d *HypervPS4Driver) verifyElevatedMode() error {

	log.Printf("Enter method: %s", "verifyElevatedMode")

	isAdmin, _ := powershell.IsCurrentUserAnAdministrator()

	if !isAdmin {
		err := fmt.Errorf("%s", "Please restart your shell in elevated mode")
		return err
	}

	return nil
}
