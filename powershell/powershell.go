// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package powershell

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	powerShellFalse = "False"
	powerShellTrue  = "True"
)

type PowerShellCmd struct {
	Stdout io.Writer
	Stderr io.Writer
}

func (ps *PowerShellCmd) Run(fileContents string, params ...string) error {
	_, err := ps.Output(fileContents, params...)
	return err
}

// Output runs the PowerShell command and returns its standard output.
func (ps *PowerShellCmd) Output(fileContents string, params ...string) (string, error) {
	path, err := ps.getPowerShellPath()
	if err != nil {
		return "", err
	}

	filename, err := saveScript(fileContents)
	if err != nil {
		return "", err
	}

	debug := os.Getenv("PACKER_POWERSHELL_DEBUG") != ""
	verbose := debug || os.Getenv("PACKER_POWERSHELL_VERBOSE") != ""

	if !debug {
		defer os.Remove(filename)
	}

	args := createArgs(filename, params...)

	if verbose {
		log.Printf("Run: %s %s", path, args)
	}

	var stdout, stderr bytes.Buffer
	command := exec.Command(path, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	err = command.Run()

	if ps.Stdout != nil {
		stdout.WriteTo(ps.Stdout)
	}

	if ps.Stderr != nil {
		stderr.WriteTo(ps.Stderr)
	}

	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("PowerShell error: %s", stderrString)
	}

	if len(stderrString) > 0 {
		err = fmt.Errorf("PowerShell error: %s", stderrString)
	}

	stdoutString := strings.TrimSpace(stdout.String())

	if verbose && stdoutString != "" {
		log.Printf("stdout: %s", stdoutString)
	}

	// only write the stderr string if verbose because
	// the error string will already be in the err return value.
	if verbose && stderrString != "" {
		log.Printf("stderr: %s", stderrString)
	}

	return stdoutString, err
}

func IsPowershellAvailable() (bool, string, error) {
	path, err := exec.LookPath("powershell")
	if err != nil {
		return false, "", err
	} else {
		return true, path, err
	}
}

func (ps *PowerShellCmd) getPowerShellPath() (string, error) {
	powershellAvailable, path, err := IsPowershellAvailable()

	if !powershellAvailable {
		log.Fatal("Cannot find PowerShell in the path")
		return "", err
	}

	return path, nil
}

func saveScript(fileContents string) (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "ps")
	if err != nil {
		return "", err
	}

	_, err = file.Write([]byte(fileContents))
	if err != nil {
		return "", err
	}

	err = file.Close()
	if err != nil {
		return "", err
	}

	newFilename := file.Name() + ".ps1"
	err = os.Rename(file.Name(), newFilename)
	if err != nil {
		return "", err
	}

	return newFilename, nil
}

func createArgs(filename string, params ...string) []string {
	args := make([]string, len(params)+4)
	args[0] = "-ExecutionPolicy"
	args[1] = "Bypass"

	args[2] = "-File"
	args[3] = filename

	for key, value := range params {
		args[key+4] = value
	}

	return args
}

func GetHostAvailableMemory() float64 {

	var script = "(Get-WmiObject Win32_OperatingSystem).FreePhysicalMemory / 1024"

	var ps PowerShellCmd
	output, _ := ps.Output(script)

	freeMB, _ := strconv.ParseFloat(output, 64)

	return freeMB
}

func GetHostName(ip string) (string, error) {

	var script = `
param([string]$ip)
try {
  $HostName = [System.Net.Dns]::GetHostEntry($ip).HostName
  if ($HostName -ne $null) {
    $HostName = $HostName.Split('.')[0]
  }
  $HostName
} catch { }
`

	//
	var ps PowerShellCmd
	cmdOut, err := ps.Output(script, ip)
	if err != nil {
		return "", err
	}

	return cmdOut, nil
}

func IsCurrentUserAnAdministrator() (bool, error) {
	var script = `
$identity = [System.Security.Principal.WindowsIdentity]::GetCurrent()
$principal = new-object System.Security.Principal.WindowsPrincipal($identity)
$administratorRole = [System.Security.Principal.WindowsBuiltInRole]::Administrator
return $principal.IsInRole($administratorRole)
`

	var ps PowerShellCmd
	cmdOut, err := ps.Output(script)
	if err != nil {
		return false, err
	}

	res := strings.TrimSpace(cmdOut)
	return res == powerShellTrue, nil
}

func ModuleExists(moduleName string) (bool, error) {

	var script = `
param([string]$moduleName)
(Get-Module -Name $moduleName) -ne $null
`
	var ps PowerShellCmd
	cmdOut, err := ps.Output(script)
	if err != nil {
		return false, err
	}

	res := strings.TrimSpace(string(cmdOut))

	if res == powerShellFalse {
		err := fmt.Errorf("PowerShell %s module is not loaded. Make sure %s feature is on.", moduleName, moduleName)
		return false, err
	}

	return true, nil
}

func HasVirtualMachineVirtualizationExtensions() (bool, error) {

	var script = `	
(GET-Command Set-VMProcessor).parameters.keys -contains "ExposeVirtualizationExtensions"
`

	var ps PowerShellCmd
	cmdOut, err := ps.Output(script)

	if err != nil {
		return false, err
	}

	var hasVirtualMachineVirtualizationExtensions = strings.TrimSpace(cmdOut) == "True"
	return hasVirtualMachineVirtualizationExtensions, err
}

func SetUnattendedProductKey(path string, productKey string) error {

	var script = `
param([string]$path,[string]$productKey)

$unattend = [xml](Get-Content -Path $path)
$ns = @{ un = 'urn:schemas-microsoft-com:unattend' }

$setupNode = $unattend | 
  Select-Xml -XPath '//un:settings[@pass = "specialize"]/un:component[@name = "Microsoft-Windows-Shell-Setup"]' -Namespace $ns |
  Select-Object -ExpandProperty Node

$productKeyNode = $setupNode |
  Select-Xml -XPath '//un:ProductKey' -Namespace $ns |
  Select-Object -ExpandProperty Node

if ($productKeyNode -eq $null) {
    $productKeyNode = $unattend.CreateElement('ProductKey', $ns.un)
    [Void]$setupNode.AppendChild($productKeyNode)
}

$productKeyNode.InnerText = $productKey

$unattend.Save($path)
`

	var ps PowerShellCmd
	err := ps.Run(script, path, productKey)
	return err
}
