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
