// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"runtime"
	"strconv"
	"bytes"
)

type HypervPS4Driver struct {
	HypervManagePath string
}

func NewHypervPS4Driver() (Driver, error) {
	appliesTo := "Applies to Windows 8.1, Windows PowerShell 4.0, Windows Server 2012 R2 only"

	// Check this is Windows
	if runtime.GOOS != "windows" {
		err := fmt.Errorf("%s", appliesTo)
		return nil, err
	}

	ps4Driver := &HypervPS4Driver{ HypervManagePath: "powershell"}

	if err := ps4Driver.Verify(); err != nil {
		return nil, err
	}

	log.Printf("HypervManage path: %s", ps4Driver.HypervManagePath)

	return ps4Driver, nil
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

	if err := d.setExecutionPolicy(); err != nil {
		return err
	}

	return nil
}

func (d *HypervPS4Driver) verifyPSVersion() error {

	log.Printf("Enter method: %s", "verifyPSVersion")
	// check PS is available and is of proper version
	versionCmd := "$host.version.Major"
	cmd := exec.Command(d.HypervManagePath, versionCmd)

	cmdOut, err := cmd.Output()
	if err != nil {
		return err
	}

	versionOutput := strings.TrimSpace(string(cmdOut))
	log.Printf("%s output: %s", versionCmd, versionOutput)

	ver, err := strconv.ParseInt(versionOutput, 10, 32)

	if  err != nil {
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

	versionCmd := "Invoke-Command -scriptblock { function foo(){try{ $commands = Get-Command -Module Hyper-V;if($commands.Length -eq 0){return $false} }catch{return $false}; return $true} foo}"
	cmd := exec.Command(d.HypervManagePath, versionCmd)

	cmdOut, err := cmd.Output()
	if err != nil {
		return err
	}

	res := strings.TrimSpace(string(cmdOut))

	if(res== "False"){
		err := fmt.Errorf("%s", "PS Hyper-V module is not loaded. Make sure Hyper-V feature is on.")
		return err
	}

	return nil
}

func (d *HypervPS4Driver) verifyElevatedMode() error {

	log.Printf("Enter method: %s", "verifyElevatedMode")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {function foo(){try{")
	blockBuffer.WriteString("$myWindowsID=[System.Security.Principal.WindowsIdentity]::GetCurrent();")
	blockBuffer.WriteString("$myWindowsPrincipal=new-object System.Security.Principal.WindowsPrincipal($myWindowsID);")
	blockBuffer.WriteString("$adminRole=[System.Security.Principal.WindowsBuiltInRole]::Administrator;")
	blockBuffer.WriteString("if($myWindowsPrincipal.IsInRole($adminRole)){return $true}else{return $false}")
	blockBuffer.WriteString("}catch{return $false}} foo}")

	log.Printf(" blockBuffer: %s", blockBuffer.String())
	cmd := exec.Command(d.HypervManagePath, blockBuffer.String())

	cmdOut, err := cmd.Output()
	if err != nil {
		return err
	}

	res := strings.TrimSpace(string(cmdOut))
	log.Printf("cmdOut: " + string(res))

	if(res == "False"){
		err := fmt.Errorf("%s", "Please restart your shell in elevated mode")
		return err
	}

	return nil
}

func (d *HypervPS4Driver) setExecutionPolicy() error {

	log.Printf("Enter method: %s", "setExecutionPolicy")

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {Set-ExecutionPolicy RemoteSigned -Force}")

	err := d.HypervManage(blockBuffer.String())

	return err
}

func (d *HypervPS4Driver) HypervManage(block string) error {

	log.Printf("Executing HypervManage: %#v", block)

	var stdout, stderr bytes.Buffer

	script := exec.Command(d.HypervManagePath, block)
	script.Stdout = &stdout
	script.Stderr = &stderr

	err := script.Run()

	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("HypervManage error: %s", stderrString)
	}

	if len(stderrString) > 0 {
		err = fmt.Errorf("HypervManage error: %s", stderrString)
	}

	stdoutString := strings.TrimSpace(stdout.String())

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return err
}
