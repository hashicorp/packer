package hyperv

import (
	"strings"
	"testing"
)

func Test_getCreateVMScript(t *testing.T) {
	vmName := "myvm"
	path := "C://mypath"
	harddrivepath := "C://harddrivepath"
	ram := int64(1024)
	disksize := int64(8192)
	diskBlockSize := int64(10)
	switchName := "hyperv-vmx-switch"
	generation := uint(1)
	diffdisks := true
	fixedVHD := true
	version := "5.0"

	// Check Fixed VHD conditional set
	scriptString, err := getCreateVMScript(vmName, path, harddrivepath, ram,
		disksize, diskBlockSize, switchName, generation, diffdisks, fixedVHD,
		version)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	expected := `$vhdPath = Join-Path -Path C://mypath -ChildPath myvm.vhd
Hyper-V\New-VHD -Path $vhdPath -ParentPath C://harddrivepath -Differencing -BlockSizeBytes 10
Hyper-V\New-VM -Name myvm -Path C://mypath -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName hyperv-vmx-switch -Version 5.0`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	// We should never get here thanks to good template validation, but it's
	// good to fail rather than trying to run the ps script and erroring.
	generation = uint(2)
	scriptString, err = getCreateVMScript(vmName, path, harddrivepath, ram,
		disksize, diskBlockSize, switchName, generation, diffdisks, fixedVHD,
		version)
	if err == nil {
		t.Fatalf("Should have Error: %s", err.Error())
	}

	// Check VHDX conditional set
	fixedVHD = false
	scriptString, err = getCreateVMScript(vmName, path, harddrivepath, ram,
		disksize, diskBlockSize, switchName, generation, diffdisks, fixedVHD,
		version)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	expected = `$vhdPath = Join-Path -Path C://mypath -ChildPath myvm.vhdx
Hyper-V\New-VHD -Path $vhdPath -ParentPath C://harddrivepath -Differencing -BlockSizeBytes 10
Hyper-V\New-VM -Name myvm -Path C://mypath -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName hyperv-vmx-switch -Generation 2 -Version 5.0`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	// Check generation 1 no fixed VHD
	fixedVHD = false
	generation = uint(1)
	scriptString, err = getCreateVMScript(vmName, path, harddrivepath, ram,
		disksize, diskBlockSize, switchName, generation, diffdisks, fixedVHD,
		version)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	expected = `$vhdPath = Join-Path -Path C://mypath -ChildPath myvm.vhdx
Hyper-V\New-VHD -Path $vhdPath -ParentPath C://harddrivepath -Differencing -BlockSizeBytes 10
Hyper-V\New-VM -Name myvm -Path C://mypath -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName hyperv-vmx-switch -Version 5.0`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	// Check that we use generation one template even if generation is unset
	generation = uint(0)
	scriptString, err = getCreateVMScript(vmName, path, harddrivepath, ram,
		disksize, diskBlockSize, switchName, generation, diffdisks, fixedVHD,
		version)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	// same "expected" as above
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	version = ""
	scriptString, err = getCreateVMScript(vmName, path, harddrivepath, ram,
		disksize, diskBlockSize, switchName, generation, diffdisks, fixedVHD,
		version)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	expected = `$vhdPath = Join-Path -Path C://mypath -ChildPath myvm.vhdx
Hyper-V\New-VHD -Path $vhdPath -ParentPath C://harddrivepath -Differencing -BlockSizeBytes 10
Hyper-V\New-VM -Name myvm -Path C://mypath -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName hyperv-vmx-switch`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	diffdisks = false
	scriptString, err = getCreateVMScript(vmName, path, harddrivepath, ram,
		disksize, diskBlockSize, switchName, generation, diffdisks, fixedVHD,
		version)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	expected = `$vhdPath = Join-Path -Path C://mypath -ChildPath myvm.vhdx
Copy-Item -Path C://harddrivepath -Destination $vhdPath
Hyper-V\New-VM -Name myvm -Path C://mypath -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName hyperv-vmx-switch`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	harddrivepath = ""
	scriptString, err = getCreateVMScript(vmName, path, harddrivepath, ram,
		disksize, diskBlockSize, switchName, generation, diffdisks, fixedVHD,
		version)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	expected = `$vhdPath = Join-Path -Path C://mypath -ChildPath myvm.vhdx
Hyper-V\New-VHD -Path $vhdPath -SizeBytes 8192 -BlockSizeBytes 10
Hyper-V\New-VM -Name myvm -Path C://mypath -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName hyperv-vmx-switch`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	fixedVHD = true
	scriptString, err = getCreateVMScript(vmName, path, harddrivepath, ram,
		disksize, diskBlockSize, switchName, generation, diffdisks, fixedVHD,
		version)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	expected = `$vhdPath = Join-Path -Path C://mypath -ChildPath myvm.vhd
Hyper-V\New-VHD -Path $vhdPath -Fixed -SizeBytes 8192
Hyper-V\New-VM -Name myvm -Path C://mypath -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName hyperv-vmx-switch`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}
}
