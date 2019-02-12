package hyperv

import (
	"strings"
	"testing"
)

func Test_getCreateVMScript(t *testing.T) {
	opts := scriptOptions{
		Version:            "5.0",
		VMName:             "myvm",
		Path:               "C://mypath",
		HardDrivePath:      "C://harddrivepath",
		MemoryStartupBytes: int64(1024),
		NewVHDSizeBytes:    int64(8192),
		VHDBlockSizeBytes:  int64(10),
		SwitchName:         "hyperv-vmx-switch",
		Generation:         uint(1),
		DiffDisks:          true,
		FixedVHD:           true,
	}

	// Check Fixed VHD conditional set
	scriptString, err := getCreateVMScript(&opts)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	expected := `$vhdPath = Join-Path -Path "C://mypath" -ChildPath "myvm.vhd"
Hyper-V\New-VHD -Path $vhdPath -ParentPath "C://harddrivepath" -Differencing -BlockSizeBytes 10
Hyper-V\New-VM -Name "myvm" -Path "C://mypath" -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName "hyperv-vmx-switch" -Version 5.0`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	// We should never get here thanks to good template validation, but it's
	// good to fail rather than trying to run the ps script and erroring.
	opts.Generation = uint(2)
	scriptString, err = getCreateVMScript(&opts)
	if err == nil {
		t.Fatalf("Should have Error: %s", err.Error())
	}

	// Check VHDX conditional set
	opts.FixedVHD = false
	scriptString, err = getCreateVMScript(&opts)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	expected = `$vhdPath = Join-Path -Path "C://mypath" -ChildPath "myvm.vhdx"
Hyper-V\New-VHD -Path $vhdPath -ParentPath "C://harddrivepath" -Differencing -BlockSizeBytes 10
Hyper-V\New-VM -Name "myvm" -Path "C://mypath" -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName "hyperv-vmx-switch" -Generation 2 -Version 5.0`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	// Check generation 1 no fixed VHD
	opts.FixedVHD = false
	opts.Generation = uint(1)
	scriptString, err = getCreateVMScript(&opts)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	expected = `$vhdPath = Join-Path -Path "C://mypath" -ChildPath "myvm.vhdx"
Hyper-V\New-VHD -Path $vhdPath -ParentPath "C://harddrivepath" -Differencing -BlockSizeBytes 10
Hyper-V\New-VM -Name "myvm" -Path "C://mypath" -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName "hyperv-vmx-switch" -Version 5.0`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	// Check that we use generation one template even if generation is unset
	opts.Generation = uint(0)
	scriptString, err = getCreateVMScript(&opts)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	// same "expected" as above
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	opts.Version = ""
	scriptString, err = getCreateVMScript(&opts)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	expected = `$vhdPath = Join-Path -Path "C://mypath" -ChildPath "myvm.vhdx"
Hyper-V\New-VHD -Path $vhdPath -ParentPath "C://harddrivepath" -Differencing -BlockSizeBytes 10
Hyper-V\New-VM -Name "myvm" -Path "C://mypath" -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName "hyperv-vmx-switch"`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	opts.DiffDisks = false
	scriptString, err = getCreateVMScript(&opts)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	expected = `$vhdPath = Join-Path -Path "C://mypath" -ChildPath "myvm.vhdx"
Copy-Item -Path "C://harddrivepath" -Destination $vhdPath
Hyper-V\New-VM -Name "myvm" -Path "C://mypath" -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName "hyperv-vmx-switch"`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	opts.HardDrivePath = ""
	scriptString, err = getCreateVMScript(&opts)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	expected = `$vhdPath = Join-Path -Path "C://mypath" -ChildPath "myvm.vhdx"
Hyper-V\New-VHD -Path $vhdPath -SizeBytes 8192 -BlockSizeBytes 10
Hyper-V\New-VM -Name "myvm" -Path "C://mypath" -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName "hyperv-vmx-switch"`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}

	opts.FixedVHD = true
	scriptString, err = getCreateVMScript(&opts)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	expected = `$vhdPath = Join-Path -Path "C://mypath" -ChildPath "myvm.vhd"
Hyper-V\New-VHD -Path $vhdPath -Fixed -SizeBytes 8192
Hyper-V\New-VM -Name "myvm" -Path "C://mypath" -MemoryStartupBytes 1024 -VHDPath $vhdPath -SwitchName "hyperv-vmx-switch"`
	if ok := strings.Compare(scriptString, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, scriptString)
	}
}
