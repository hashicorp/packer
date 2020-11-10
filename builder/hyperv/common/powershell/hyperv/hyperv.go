package hyperv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/hashicorp/packer/builder/hyperv/common/powershell"
)

type scriptOptions struct {
	Version            string
	VMName             string
	VHDX               string
	Path               string
	HardDrivePath      string
	MemoryStartupBytes int64
	NewVHDSizeBytes    int64
	VHDBlockSizeBytes  int64
	SwitchName         string
	Generation         uint
	DiffDisks          bool
	FixedVHD           bool
}

func GetHostAdapterIpAddressForSwitch(switchName string) (string, error) {
	var script = `
param([string]$switchName, [int]$addressIndex)
$HostVMAdapter = Hyper-V\Get-VMNetworkAdapter -ManagementOS -SwitchName $switchName | Select-Object -First 1
if ($HostVMAdapter){
  $HostNetAdapter = Get-NetAdapter | Where-Object { $_.DeviceId -eq $HostVMAdapter.DeviceId }
  if ($HostNetAdapter){
    $HostNetAdapterIfIndex = @()
    $HostNetAdapterIfIndex += $HostNetAdapter.ifIndex
    $HostNetAdapterConfiguration = @(get-wmiobject win32_networkadapterconfiguration -filter "IPEnabled = 'TRUE'") | Where-Object { $HostNetAdapterIfIndex.Contains($_.InterfaceIndex)}
    if ($HostNetAdapterConfiguration){
      return @($HostNetAdapterConfiguration.IpAddress)[$addressIndex]
    }
  }
} else {
  $HostNetAdapterConfiguration=@(Get-NetIPAddress -CimSession $env:computername -AddressFamily IPv4 | Where-Object { ( $_.InterfaceAlias -notmatch 'Loopback' ) -and ( $_.SuffixOrigin -notmatch "Link" )})
  if ($HostNetAdapterConfiguration) {
    return @($HostNetAdapterConfiguration.IpAddress)[$addressIndex]
  } else {
    return $false
 }
}
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, switchName, "0")

	return cmdOut, err
}

func GetVirtualMachineNetworkAdapterAddress(vmName string) (string, error) {

	var script = `
param([string]$vmName, [int]$addressIndex)
try {
  $adapter = Hyper-V\Get-VMNetworkAdapter -VMName $vmName -ErrorAction SilentlyContinue
  if ($adapter.IPAddresses) {
    $ip = $adapter.IPAddresses[$addressIndex]
  } else {
    $vm = Get-CimInstance -ClassName Msvm_ComputerSystem -Namespace root\virtualization\v2 -Filter "ElementName='$vmName'"
    $ip_details = (Get-CimAssociatedInstance -InputObject $vm -ResultClassName Msvm_KvpExchangeComponent).GuestIntrinsicExchangeItems | %{ [xml]$_ } | ?{ $_.SelectSingleNode("/INSTANCE/PROPERTY[@NAME='Name']/VALUE[child::text()='NetworkAddressIPv4']") }

    if ($null -eq $ip_details) {
      return $false
    }

    $ip_addresses = $ip_details.SelectSingleNode("/INSTANCE/PROPERTY[@NAME='Data']/VALUE/child::text()").Value
    $ip = ($ip_addresses -split ";")[0]
  }
} catch {
  return $false
}
$ip
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, vmName, "0")

	return cmdOut, err
}

func CreateDvdDrive(vmName string, isoPath string, generation uint) (uint, uint, error) {
	var ps powershell.PowerShellCmd
	var script string

	script = `
param([string]$vmName, [string]$isoPath)
$dvdController = Hyper-V\Add-VMDvdDrive -VMName $vmName -path $isoPath -Passthru
$dvdController | Hyper-V\Set-VMDvdDrive -path $null
$result = "$($dvdController.ControllerNumber),$($dvdController.ControllerLocation)"
$result
`

	cmdOut, err := ps.Output(script, vmName, isoPath)
	if err != nil {
		return 0, 0, err
	}

	cmdOutArray := strings.Split(cmdOut, ",")
	if len(cmdOutArray) != 2 {
		return 0, 0, errors.New("Did not return controller number and controller location")
	}

	controllerNumberTemp, err := strconv.ParseUint(strings.TrimSpace(cmdOutArray[0]), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	controllerNumber := uint(controllerNumberTemp)

	controllerLocationTemp, err := strconv.ParseUint(strings.TrimSpace(cmdOutArray[1]), 10, 64)
	if err != nil {
		return controllerNumber, 0, err
	}
	controllerLocation := uint(controllerLocationTemp)

	return controllerNumber, controllerLocation, err
}

func MountDvdDrive(vmName string, path string, controllerNumber uint, controllerLocation uint) error {

	var script = `
param([string]$vmName,[string]$path,[string]$controllerNumber,[string]$controllerLocation)
$vmDvdDrive = Hyper-V\Get-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation
if (!$vmDvdDrive) {throw 'unable to find dvd drive'}
Hyper-V\Set-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation -Path $path
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, path, strconv.FormatInt(int64(controllerNumber), 10),
		strconv.FormatInt(int64(controllerLocation), 10))
	return err
}

func UnmountDvdDrive(vmName string, controllerNumber uint, controllerLocation uint) error {
	var script = `
param([string]$vmName,[int]$controllerNumber,[int]$controllerLocation)
$vmDvdDrive = Hyper-V\Get-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation
if (!$vmDvdDrive) {throw 'unable to find dvd drive'}
Hyper-V\Set-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation -Path $null
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, strconv.FormatInt(int64(controllerNumber), 10),
		strconv.FormatInt(int64(controllerLocation), 10))
	return err
}

func SetBootDvdDrive(vmName string, controllerNumber uint, controllerLocation uint, generation uint) error {

	if generation < 2 {
		script := `
param([string]$vmName)
Hyper-V\Set-VMBios -VMName $vmName -StartupOrder @("IDE","CD","LegacyNetworkAdapter","Floppy")
`
		var ps powershell.PowerShellCmd
		err := ps.Run(script, vmName)
		return err
	} else {
		script := `
param([string]$vmName,[int]$controllerNumber,[int]$controllerLocation)
$vmDvdDrive = Hyper-V\Get-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation
if (!$vmDvdDrive) {throw 'unable to find dvd drive'}
Hyper-V\Set-VMFirmware -VMName $vmName -FirstBootDevice $vmDvdDrive -ErrorAction SilentlyContinue
`
		var ps powershell.PowerShellCmd
		err := ps.Run(script, vmName, strconv.FormatInt(int64(controllerNumber), 10),
			strconv.FormatInt(int64(controllerLocation), 10))
		return err
	}
}

func SetFirstBootDeviceGen1(vmName string, controllerType string) error {

	// for Generation 1 VMs, we read the value of the VM's boot order, strip the value specified in
	// controllerType and insert that value back at the beginning of the list.
	//
	// controllerType must be 'NET', 'DVD', 'IDE' or 'FLOPPY' (case sensitive)
	// The 'NET' value is always replaced with 'LegacyNetworkAdapter'

	if controllerType == "NET" {
		controllerType = "LegacyNetworkAdapter"
	}

	script := `
param([string] $vmName, [string] $controllerType)
	$vmBootOrder = Hyper-V\Get-VMBios -VMName $vmName | Select-Object -ExpandProperty StartupOrder | Where-Object { $_ -ne $controllerType }
	Hyper-V\Set-VMBios -VMName $vmName -StartupOrder (@($controllerType) + $vmBootOrder)
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, controllerType)
	return err
}

func SetFirstBootDeviceGen2(vmName string, controllerType string, controllerNumber uint, controllerLocation uint) error {

	script := `param ([string] $vmName, [string] $controllerType, [int] $controllerNumber, [int] $controllerLocation)`

	switch {

	case controllerType == "CD":
		// for CDs we have to use Get-VMDvdDrive to find the device
		script += `
$vmDevice = Hyper-V\Get-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation -ErrorAction SilentlyContinue`

	case controllerType == "NET":
		// for "NET" device, we select the first network adapter on the VM
		script += `
$vmDevice = Hyper-V\Get-VMNetworkAdapter -VMName $vmName -ErrorAction SilentlyContinue | Select-Object -First 1`

	default:
		script += `
$vmDevice = @(Hyper-V\Get-VMIdeController -VMName $vmName -ErrorAction SilentlyContinue) +
	@(Hyper-V\Get-VMScsiController -VMName $vmName -ErrorAction SilentlyContinue) |
	Select-Object -ExpandProperty Drives |
	Where-Object { $_.ControllerType -eq $controllerType } |
	Where-Object { ($_.ControllerNumber -eq $controllerNumber) -and ($_.ControllerLocation -eq $controllerLocation) }
`

	}

	script += `
if ($vmDevice -eq $null) { throw 'unable to find boot device' }
Hyper-V\Set-VMFirmware -VMName $vmName -FirstBootDevice $vmDevice
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, controllerType, strconv.FormatInt(int64(controllerNumber), 10), strconv.FormatInt(int64(controllerLocation), 10))
	return err
}

func SetFirstBootDevice(vmName string, controllerType string, controllerNumber uint, controllerLocation uint, generation uint) error {

	if generation == 1 {
		return SetFirstBootDeviceGen1(vmName, controllerType)
	} else {
		return SetFirstBootDeviceGen2(vmName, controllerType, controllerNumber, controllerLocation)
	}
}

func SetBootOrder(vmName string, bootOrder []string) error {
	var script = `
param([string]$vmName, [Parameter(ValueFromRemainingArguments=$true)]$bootOrder)

$bootOrderDrives = $bootOrder | ForEach-Object {
	if ($_ -match 'SCSI:([0-9]+):([0-9]+)') {
		$controllerNumber = $Matches[1]
		$controllerLocation = $Matches[2]
		$controller = Hyper-V\Get-VMScsiController -ControllerNumber $controllerNumber $vmName
		$controller.Drives | Where-Object {$_.ControllerLocation -eq $controllerLocation} | Select-Object -First 1
	}
}

Hyper-V\Set-VMFirmware $vmName -BootOrder $bootOrderDrives
`
	var ps powershell.PowerShellCmd
	params := append([]string{vmName}, bootOrder...)
	err := ps.Run(script, params...)
	return err
}

func DeleteDvdDrive(vmName string, controllerNumber uint, controllerLocation uint) error {
	var script = `
param([string]$vmName,[int]$controllerNumber,[int]$controllerLocation)
$vmDvdDrive = Hyper-V\Get-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation
if (!$vmDvdDrive) {throw 'unable to find dvd drive'}
Hyper-V\Remove-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, strconv.FormatInt(int64(controllerNumber), 10),
		strconv.FormatInt(int64(controllerLocation), 10))
	return err
}

func DeleteAllDvdDrives(vmName string) error {
	var script = `
param([string]$vmName)
Hyper-V\Get-VMDvdDrive -VMName $vmName | Hyper-V\Remove-VMDvdDrive
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func MountFloppyDrive(vmName string, path string) error {
	var script = `
param([string]$vmName, [string]$path)
Hyper-V\Set-VMFloppyDiskDrive -VMName $vmName -Path $path
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, path)
	return err
}

func UnmountFloppyDrive(vmName string) error {

	var script = `
param([string]$vmName)
Hyper-V\Set-VMFloppyDiskDrive -VMName $vmName -Path $null
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

// This was created as a proof of concept for moving logic out of the powershell
// scripting so that we can test the pathways activated by our variables.
// Rather than creating a powershell script with several conditionals, this will
// generate a conditional-free script which already sets all of the necessary
// variables inline.
//
// For examples of what this template will generate, you can look at the
// test cases in ./hyperv_test.go
//
func getCreateVMScript(opts *scriptOptions) (string, error) {

	if opts.FixedVHD && opts.Generation == 2 {
		return "", fmt.Errorf("Generation 2 VMs don't support fixed disks.")
	}

	opts.VHDX = opts.VMName + ".vhdx"
	if opts.FixedVHD {
		opts.VHDX = opts.VMName + ".vhd"
	}

	var tpl = template.Must(template.New("createVM").Parse(`
$vhdPath = Join-Path -Path "{{ .Path }}" -ChildPath "{{ .VHDX }}"

{{ if ne .HardDrivePath "" -}}
    {{- if .DiffDisks -}}
    Hyper-V\New-VHD -Path $vhdPath -ParentPath "{{ .HardDrivePath }}" -Differencing -BlockSizeBytes {{ .VHDBlockSizeBytes }}
    {{- else -}}
    Copy-Item -Path "{{ .HardDrivePath }}" -Destination $vhdPath
    {{- end -}}
{{- else -}}
    {{- if .FixedVHD -}}
    Hyper-V\New-VHD -Path $vhdPath -Fixed -SizeBytes {{ .NewVHDSizeBytes }}
    {{- else -}}
    Hyper-V\New-VHD -Path $vhdPath -SizeBytes {{ .NewVHDSizeBytes }} -BlockSizeBytes {{ .VHDBlockSizeBytes }}
    {{- end -}}
{{- end }}

Hyper-V\New-VM -Name "{{ .VMName }}" -Path "{{ .Path }}" -MemoryStartupBytes {{ .MemoryStartupBytes }} -VHDPath $vhdPath -SwitchName "{{ .SwitchName }}"
{{- if eq .Generation 2}} -Generation {{ .Generation }} {{- end -}}
{{- if ne .Version ""}} -Version {{ .Version }} {{- end -}}
`))

	var b bytes.Buffer
	err := tpl.Execute(&b, opts)
	if err != nil {
		return "", err
	}

	// Tidy away the excess newlines left over by the template
	regex, err := regexp.Compile("^\n")
	if err != nil {
		return "", err
	}
	final := regex.ReplaceAllString(b.String(), "")
	regex, err = regexp.Compile("\n\n")
	if err != nil {
		return "", err
	}
	final = regex.ReplaceAllString(final, "\n")

	return final, nil
}

func CheckVMName(vmName string) error {
	// Check that no vm with the same name is registered, to prevent
	// namespace collisions
	var gs powershell.PowerShellCmd
	getVMCmd := fmt.Sprintf(`Hyper-V\Get-VM -Name "%s"`, vmName)
	if err := gs.Run(getVMCmd); err == nil {
		return fmt.Errorf("A virtual machine with the name %s is already"+
			" defined in Hyper-V. To avoid a name collision, please set your "+
			"vm_name to a unique value", vmName)
	}

	return nil
}

func CreateVirtualMachine(vmName string, path string, harddrivePath string, ram int64,
	diskSize int64, diskBlockSize int64, switchName string, generation uint,
	diffDisks bool, fixedVHD bool, version string) error {
	opts := scriptOptions{
		Version:            version,
		VMName:             vmName,
		Path:               path,
		HardDrivePath:      harddrivePath,
		MemoryStartupBytes: ram,
		NewVHDSizeBytes:    diskSize,
		VHDBlockSizeBytes:  diskBlockSize,
		SwitchName:         switchName,
		Generation:         generation,
		DiffDisks:          diffDisks,
		FixedVHD:           fixedVHD,
	}

	script, err := getCreateVMScript(&opts)
	if err != nil {
		return err
	}

	var ps powershell.PowerShellCmd
	if err = ps.Run(script); err != nil {
		return err
	}

	if err := DisableAutomaticCheckpoints(vmName); err != nil {
		return err
	}
	if generation != 2 {
		return DeleteAllDvdDrives(vmName)
	}
	return nil
}

func DisableAutomaticCheckpoints(vmName string) error {
	var script = `
param([string]$vmName)
if ((Get-Command Hyper-V\Set-Vm).Parameters["AutomaticCheckpointsEnabled"]) {
	Hyper-V\Set-Vm -Name $vmName -AutomaticCheckpointsEnabled $false }
`
	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func ExportVmcxVirtualMachine(exportPath string, vmName string, snapshotName string, allSnapshots bool) error {
	var script = `
param([string]$exportPath, [string]$vmName, [string]$snapshotName, [string]$allSnapshotsString)

$WorkingPath = Join-Path $exportPath $vmName

if (Test-Path $WorkingPath) {
	throw "Export path working directory: $WorkingPath already exists!"
}

$allSnapshots = [System.Boolean]::Parse($allSnapshotsString)

if ($snapshotName) {
    $snapshot = Hyper-V\Get-VMSnapshot -VMName $vmName -Name $snapshotName
    Hyper-V\Export-VMSnapshot -VMSnapshot $snapshot -Path $exportPath -ErrorAction Stop
} else {
    if (!$allSnapshots) {
        #Use last snapshot if one was not specified
        $snapshot = Hyper-V\Get-VMSnapshot -VMName $vmName | Select -Last 1
    } else {
        $snapshot = $null
    }

    if (!$snapshot) {
        #No snapshot clone
        Hyper-V\Export-VM -Name $vmName -Path $exportPath -ErrorAction Stop
    } else {
        #Snapshot clone
        Hyper-V\Export-VMSnapshot -VMSnapshot $snapshot -Path $exportPath -ErrorAction Stop
    }
}

$result = Get-ChildItem -Path $WorkingPath | Move-Item -Destination $exportPath -Force
$result = Remove-Item -Path $WorkingPath
	`

	allSnapshotsString := "False"
	if allSnapshots {
		allSnapshotsString = "True"
	}

	var ps powershell.PowerShellCmd
	err := ps.Run(script, exportPath, vmName, snapshotName, allSnapshotsString)

	return err
}

func CopyVmcxVirtualMachine(exportPath string, cloneFromVmcxPath string) error {
	var script = `
param([string]$exportPath, [string]$cloneFromVmcxPath)
if (!(Test-Path $cloneFromVmcxPath)){
	throw "Clone from vmcx directory: $cloneFromVmcxPath does not exist!"
}

if (!(Test-Path $exportPath)){
	New-Item -ItemType Directory -Force -Path $exportPath
}
$cloneFromVmcxPath = Join-Path $cloneFromVmcxPath '\*'
Copy-Item $cloneFromVmcxPath $exportPath -Recurse -Force
	`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, exportPath, cloneFromVmcxPath)

	return err
}

func SetVmNetworkAdapterMacAddress(vmName string, mac string) error {
	var script = `
param([string]$vmName, [string]$mac)
Hyper-V\Set-VMNetworkAdapter $vmName -staticmacaddress $mac
	`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, mac)

	return err
}

func ImportVmcxVirtualMachine(importPath string, vmName string, harddrivePath string,
	ram int64, switchName string, copyTF bool) error {

	var script = `
param([string]$importPath, [string]$vmName, [string]$harddrivePath, [long]$memoryStartupBytes, [string]$switchName, [string]$copy)

$VirtualHarddisksPath = Join-Path -Path $importPath -ChildPath 'Virtual Hard Disks'
if (!(Test-Path $VirtualHarddisksPath)) {
	New-Item -ItemType Directory -Force -Path $VirtualHarddisksPath
}

$vhdPath = ""
if ($harddrivePath){
	$vhdx = $vmName + '.vhdx'
	$vhdPath = Join-Path -Path $VirtualHarddisksPath -ChildPath $vhdx
}

$VirtualMachinesPath = Join-Path $importPath 'Virtual Machines'
if (!(Test-Path $VirtualMachinesPath)) {
	New-Item -ItemType Directory -Force -Path $VirtualMachinesPath
}

$VirtualMachinePath = Get-ChildItem -Path $VirtualMachinesPath -Filter *.vmcx -Recurse -ErrorAction SilentlyContinue | select -First 1 | %{$_.FullName}
if (!$VirtualMachinePath){
    $VirtualMachinePath = Get-ChildItem -Path $VirtualMachinesPath -Filter *.xml -Recurse -ErrorAction SilentlyContinue | select -First 1 | %{$_.FullName}
}
if (!$VirtualMachinePath){
    $VirtualMachinePath = Get-ChildItem -Path $importPath -Filter *.xml -Recurse -ErrorAction SilentlyContinue | select -First 1 | %{$_.FullName}
}

$copyBool = $false
switch($copy) {
    "true" { $copyBool = $true }
    default { $copyBool = $false }
}

$compatibilityReport = Hyper-V\Compare-VM -Path $VirtualMachinePath -VirtualMachinePath $importPath -SmartPagingFilePath $importPath -SnapshotFilePath $importPath -VhdDestinationPath $VirtualHarddisksPath -GenerateNewId -Copy:$false
if ($vhdPath){
	Copy-Item -Path $harddrivePath -Destination $vhdPath
	$existingFirstHarddrive = $compatibilityReport.VM.HardDrives | Select -First 1
	if ($existingFirstHarddrive) {
		$existingFirstHarddrive | Hyper-V\Set-VMHardDiskDrive -Path $vhdPath
	} else {
		Hyper-V\Add-VMHardDiskDrive -VM $compatibilityReport.VM -Path $vhdPath
	}
}
Hyper-V\Set-VMMemory -VM $compatibilityReport.VM -StartupBytes $memoryStartupBytes
$networkAdaptor = $compatibilityReport.VM.NetworkAdapters | Select -First 1
Hyper-V\Disconnect-VMNetworkAdapter -VMNetworkAdapter $networkAdaptor
Hyper-V\Connect-VMNetworkAdapter -VMNetworkAdapter $networkAdaptor -SwitchName $switchName
$vm = Hyper-V\Import-VM -CompatibilityReport $compatibilityReport

if ($vm) {
    $result = Hyper-V\Rename-VM -VM $vm -NewName $VMName
}
	`
	var ps powershell.PowerShellCmd
	err := ps.Run(script, importPath, vmName, harddrivePath, strconv.FormatInt(ram, 10), switchName, strconv.FormatBool(copyTF))

	return err
}

func CloneVirtualMachine(cloneFromVmcxPath string, cloneFromVmName string,
	cloneFromSnapshotName string, cloneAllSnapshots bool, vmName string,
	path string, harddrivePath string, ram int64, switchName string, copyTF bool) error {

	if cloneFromVmName != "" {
		if err := ExportVmcxVirtualMachine(path, cloneFromVmName,
			cloneFromSnapshotName, cloneAllSnapshots); err != nil {
			return err
		}
	}

	if cloneFromVmcxPath != "" {
		if err := CopyVmcxVirtualMachine(path, cloneFromVmcxPath); err != nil {
			return err
		}
	}

	if err := ImportVmcxVirtualMachine(path, vmName, harddrivePath, ram, switchName, copyTF); err != nil {
		return err
	}

	return DeleteAllDvdDrives(vmName)
}

func GetVirtualMachineGeneration(vmName string) (uint, error) {
	var script = `
param([string]$vmName)
$generation = Hyper-V\Get-Vm -Name $vmName | %{$_.Generation}
if (!$generation){
    $generation = 1
}
return $generation
`
	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, vmName)

	if err != nil {
		return 0, err
	}

	generationUint32, err := strconv.ParseUint(strings.TrimSpace(string(cmdOut)), 10, 32)

	if err != nil {
		return 0, err
	}

	generation := uint(generationUint32)

	return generation, err
}

func SetVirtualMachineCpuCount(vmName string, cpu uint) error {

	var script = `
param([string]$vmName, [int]$cpu)
Hyper-V\Set-VMProcessor -VMName $vmName -Count $cpu
`
	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, strconv.FormatInt(int64(cpu), 10))
	return err
}

func SetVirtualMachineVirtualizationExtensions(vmName string, enableVirtualizationExtensions bool) error {

	var script = `
param([string]$vmName, [string]$exposeVirtualizationExtensionsString)
$exposeVirtualizationExtensions = [System.Boolean]::Parse($exposeVirtualizationExtensionsString)
Hyper-V\Set-VMProcessor -VMName $vmName -ExposeVirtualizationExtensions $exposeVirtualizationExtensions
`
	exposeVirtualizationExtensionsString := "False"
	if enableVirtualizationExtensions {
		exposeVirtualizationExtensionsString = "True"
	}
	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, exposeVirtualizationExtensionsString)
	return err
}

func SetVirtualMachineDynamicMemory(vmName string, enableDynamicMemory bool) error {

	var script = `
param([string]$vmName, [string]$enableDynamicMemoryString)
$enableDynamicMemory = [System.Boolean]::Parse($enableDynamicMemoryString)
Hyper-V\Set-VMMemory -VMName $vmName -DynamicMemoryEnabled $enableDynamicMemory
`
	enableDynamicMemoryString := "False"
	if enableDynamicMemory {
		enableDynamicMemoryString = "True"
	}
	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, enableDynamicMemoryString)
	return err
}

func SetVirtualMachineMacSpoofing(vmName string, enableMacSpoofing bool) error {
	var script = `
param([string]$vmName, $enableMacSpoofing)
Hyper-V\Set-VMNetworkAdapter -VMName $vmName -MacAddressSpoofing $enableMacSpoofing
`

	var ps powershell.PowerShellCmd

	enableMacSpoofingString := "Off"
	if enableMacSpoofing {
		enableMacSpoofingString = "On"
	}

	err := ps.Run(script, vmName, enableMacSpoofingString)
	return err
}

func SetVirtualMachineSecureBoot(vmName string, enableSecureBoot bool, templateName string) error {
	var script = `
param([string]$vmName, [string]$enableSecureBootString, [string]$templateName)
$cmdlet = Get-Command Hyper-V\Set-VMFirmware
# The SecureBootTemplate parameter is only available in later versions
if ($cmdlet.Parameters.SecureBootTemplate) {
	Hyper-V\Set-VMFirmware -VMName $vmName -EnableSecureBoot $enableSecureBootString -SecureBootTemplate $templateName
} else {
	Hyper-V\Set-VMFirmware -VMName $vmName -EnableSecureBoot $enableSecureBootString
}
`

	var ps powershell.PowerShellCmd

	enableSecureBootString := "Off"
	if enableSecureBoot {
		enableSecureBootString = "On"
	}

	if templateName == "" {
		templateName = "MicrosoftWindows"
	}

	err := ps.Run(script, vmName, enableSecureBootString, templateName)
	return err
}

func DeleteVirtualMachine(vmName string) error {

	var script = `
param([string]$vmName)

$vm = Hyper-V\Get-VM -Name $vmName
if (($vm.State -ne [Microsoft.HyperV.PowerShell.VMState]::Off) -and ($vm.State -ne [Microsoft.HyperV.PowerShell.VMState]::OffCritical)) {
    Hyper-V\Stop-VM -VM $vm -TurnOff -Force -Confirm:$false
}

Hyper-V\Remove-VM -Name $vmName -Force -Confirm:$false
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func ExportVirtualMachine(vmName string, path string) error {

	var script = `
param([string]$vmName, [string]$path)
Hyper-V\Export-VM -Name $vmName -Path $path

if (Test-Path -Path ([IO.Path]::Combine($path, $vmName, 'Virtual Machines', '*.VMCX')))
{
  $vm = Hyper-V\Get-VM -Name $vmName
  $vm_adapter = Hyper-V\Get-VMNetworkAdapter -VM $vm | Select -First 1

  $config = [xml]@"
<?xml version="1.0" ?>
<configuration>
  <properties>
    <subtype type="integer">$($vm.Generation - 1)</subtype>
    <name type="string">$($vm.Name)</name>
  </properties>
  <settings>
    <processors>
      <count type="integer">$($vm.ProcessorCount)</count>
    </processors>
    <memory>
      <bank>
        <dynamic_memory_enabled type="bool">$($vm.DynamicMemoryEnabled)</dynamic_memory_enabled>
        <limit type="integer">$($vm.MemoryMaximum / 1MB)</limit>
        <reservation type="integer">$($vm.MemoryMinimum / 1MB)</reservation>
        <size type="integer">$($vm.MemoryStartup / 1MB)</size>
      </bank>
    </memory>
  </settings>
  <AltSwitchName type="string">$($vm_adapter.SwitchName)</AltSwitchName>
  <boot>
    <device0 type="string">Optical</device0>
  </boot>
  <secure_boot_enabled type="bool">False</secure_boot_enabled>
  <secure_boot_template type="string">MicrosoftWindows</secure_boot_template>
  <notes type="string">$($vm.Notes)</notes>
  <vm-controllers/>
</configuration>
"@

  if ($vm.Generation -eq 1)
  {
    $vm_controllers  = Hyper-V\Get-VMIdeController -VM $vm
    $controller_type = $config.SelectSingleNode('/configuration/vm-controllers')
    # IDE controllers are not stored in a special XML container
  }
  else
  {
    $vm_controllers  = Hyper-V\Get-VMScsiController -VM $vm
    $controller_type = $config.CreateElement('scsi')
    $controller_type.SetAttribute('ChannelInstanceGuid', 'x')
    # SCSI controllers are stored in the scsi XML container
    if ((Hyper-V\Get-VMFirmware -VM $vm).SecureBoot -eq [Microsoft.HyperV.PowerShell.OnOffState]::On)
    {
	  $config.configuration.secure_boot_enabled.'#text' = 'True'
	  $config.configuration.secure_boot_template.'#text' = (Hyper-V\Get-VMFirmware -VM $vm).SecureBootTemplate
	}
    else
    {
      $config.configuration.secure_boot_enabled.'#text' = 'False'
	}
  }

  $vm_controllers | ForEach {
    $controller = $config.CreateElement('controller' + $_.ControllerNumber)
    $_.Drives | ForEach {
      $drive = $config.CreateElement('drive' + ($_.DiskNumber + 0))
      $drive_path = $config.CreateElement('pathname')
      $drive_path.SetAttribute('type', 'string')
      $drive_path.AppendChild($config.CreateTextNode($_.Path))
      $drive_type = $config.CreateElement('type')
      $drive_type.SetAttribute('type', 'string')
      if ($_ -is [Microsoft.HyperV.PowerShell.HardDiskDrive])
      {
        $drive_type.AppendChild($config.CreateTextNode('VHD'))
      }
      elseif ($_ -is [Microsoft.HyperV.PowerShell.DvdDrive])
      {
        $drive_type.AppendChild($config.CreateTextNode('ISO'))
      }
      else
      {
        $drive_type.AppendChild($config.CreateTextNode('NONE'))
      }
      $drive.AppendChild($drive_path)
      $drive.AppendChild($drive_type)
      $controller.AppendChild($drive)
    }
    $controller_type.AppendChild($controller)
  }
  if ($controller_type.Name -ne 'vm-controllers')
  {
    $config.SelectSingleNode('/configuration/vm-controllers').AppendChild($controller_type)
  }

  $config.Save([IO.Path]::Combine($path, $vm.Name, 'Virtual Machines', 'box.xml'))
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, path)
	return err
}

func PreserveLegacyExportBehaviour(srcPath, dstPath string) error {

	var script = `
param([string]$srcPath, [string]$dstPath)

# Validate the paths returning an error if they are empty or don't exist
$srcPath, $dstPath | % {
    if ($_) {
        if (! (Test-Path $_)) {
            [System.Console]::Error.WriteLine("Path $_ does not exist")
            exit
        }
    } else {
        [System.Console]::Error.WriteLine("A supplied path is empty")
        exit
    }
}

# Export-VM should just create directories at the root of the export path
# but, just in case, move all files as well...
Move-Item -Path (Join-Path (Get-Item $srcPath).FullName "*.*") -Destination (Get-Item $dstPath).FullName

# Move directories with content; Delete empty directories
$dirObj = Get-ChildItem $srcPath -Directory | % {
    New-Object PSObject -Property @{
        FullName=$_.FullName;
        HasContent=$(if ($_.GetFileSystemInfos().Count -gt 0) {$true} else {$false})
    }
}
foreach ($directory in $dirObj) {
    if ($directory.HasContent) {
        Move-Item -Path $directory.FullName -Destination (Get-Item $dstPath).FullName
    } else {
        Remove-Item -Path $directory.FullName
    }
}

# Only remove the source directory if it is now empty
if ( $((Get-Item $srcPath).GetFileSystemInfos().Count) -eq 0 ) {
    Remove-Item -Path $srcPath
} else {
    # 'Return' an error message to PowerShellCmd as the directory should
    # always be empty at the end of the script. The check is here to stop
    # the Remove-Item command from doing any damage if some unforeseen
    # error has occured
    [System.Console]::Error.WriteLine("Refusing to remove $srcPath as it is not empty")
    exit
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, srcPath, dstPath)

	return err
}

func MoveCreatedVHDsToOutputDir(srcPath, dstPath string) error {

	var script = `
param([string]$srcPath, [string]$dstPath)

# Validate the paths returning an error if the supplied path is empty
# or if the paths don't exist
$srcPath, $dstPath | % {
    if ($_) {
        if (! (Test-Path $_)) {
            [System.Console]::Error.WriteLine("Path $_ does not exist")
            exit
        }
    } else {
        [System.Console]::Error.WriteLine("A supplied path is empty")
        exit
    }
}

# Convert to absolute paths if required
$srcPathAbs = (Get-Item($srcPath)).FullName
$dstPathAbs = (Get-Item($dstPath)).FullName

# Get the full path to all disks under the directory or exit if none are found
$disks = Get-ChildItem -Path $srcPathAbs -Recurse -Filter *.vhd* -ErrorAction SilentlyContinue | % { $_.FullName }
if ($disks.Length -eq 0) {
    [System.Console]::Error.WriteLine("No disks found under $srcPathAbs")
    exit
}

# Set up directory for VHDs in the destination directory
$vhdDstDir = Join-Path -Path $dstPathAbs -ChildPath 'Virtual Hard Disks'
if (! (Test-Path $vhdDstDir)) {
	New-Item -ItemType Directory -Force -Path $vhdDstDir
}

# Move the disks
foreach ($disk in $disks) {
	Move-Item -Path $disk -Destination $vhdDstDir
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, srcPath, dstPath)

	return err
}

func CompactDisks(path string) (result string, err error) {
	var script = `
param([string]$srcPath)

$disks = Get-ChildItem -Path $srcPath -Recurse -ErrorAction SilentlyContinue |where {$_.extension -in ".vhdx",".vhd"} |foreach { $_.FullName }
# Failure to find any disks is treated as a 'soft' error. Simply print out
# a warning and exit
if ($disks.Length -eq 0) {
    Write-Output "WARNING: No disks found under $srcPath"
    exit
}

foreach ($disk in $disks) {
    Write-Output "Compacting disk: $(Split-Path $disk -leaf)"

    $sizeBefore = (Get-Item -Path $disk).Length
    Optimize-VHD -Path $disk -Mode Full
    $sizeAfter = (Get-Item -Path $disk).Length

    # Calculate the percentage change in disk size
    if ($sizeAfter -gt 0) { # Protect against division by zero
        $percentChange = ( ( $sizeAfter / $sizeBefore ) * 100 ) - 100
        switch($percentChange) {
            {$_ -lt 0} {Write-Output "Disk size reduced by: $(([math]::Abs($_)).ToString("#.#"))%"}
            {$_ -eq 0} {Write-Output "Disk size is unchanged"}
            {$_ -gt 0} {Write-Output "WARNING: Disk size increased by: $($_.ToString("#.#"))%"}
        }
    }
}
`

	var ps powershell.PowerShellCmd
	result, err = ps.Output(script, path)
	return
}

func CreateVirtualSwitch(switchName string, switchType string) (bool, error) {

	var script = `
param([string]$switchName,[string]$switchType)
$switches = Hyper-V\Get-VMSwitch -Name $switchName -ErrorAction SilentlyContinue
if ($switches.Count -eq 0) {
  Hyper-V\New-VMSwitch -Name $switchName -SwitchType $switchType
  return $true
}
return $false
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, switchName, switchType)
	var created = strings.TrimSpace(cmdOut) == "True"
	return created, err
}

func DeleteVirtualSwitch(switchName string) error {

	var script = `
param([string]$switchName)
$switch = Hyper-V\Get-VMSwitch -Name $switchName -ErrorAction SilentlyContinue
if ($switch -ne $null) {
    $switch | Hyper-V\Remove-VMSwitch -Force -Confirm:$false
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, switchName)
	return err
}

func StartVirtualMachine(vmName string) error {

	var script = `
param([string]$vmName)
$vm = Hyper-V\Get-VM -Name $vmName -ErrorAction SilentlyContinue
if ($vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Off) {
  Hyper-V\Start-VM -Name $vmName -Confirm:$false
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func RestartVirtualMachine(vmName string) error {

	var script = `
param([string]$vmName)
Hyper-V\Restart-VM $vmName -Force -Confirm:$false
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func StopVirtualMachine(vmName string) error {

	var script = `
param([string]$vmName)
$vm = Hyper-V\Get-VM -Name $vmName
if ($vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Running) {
    Hyper-V\Stop-VM -VM $vm -Force -Confirm:$false
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func EnableVirtualMachineIntegrationService(vmName string, integrationServiceName string) error {

	integrationServiceId := ""
	switch integrationServiceName {
	case "Time Synchronization":
		integrationServiceId = "2497F4DE-E9FA-4204-80E4-4B75C46419C0"
	case "Heartbeat":
		integrationServiceId = "84EAAE65-2F2E-45F5-9BB5-0E857DC8EB47"
	case "Key-Value Pair Exchange":
		integrationServiceId = "2A34B1C2-FD73-4043-8A5B-DD2159BC743F"
	case "Shutdown":
		integrationServiceId = "9F8233AC-BE49-4C79-8EE3-E7E1985B2077"
	case "VSS":
		integrationServiceId = "5CED1297-4598-4915-A5FC-AD21BB4D02A4"
	case "Guest Service Interface":
		integrationServiceId = "6C09BB55-D683-4DA0-8931-C9BF705F6480"
	default:
		panic("unrecognized Integration Service Name")
	}

	var script = `
param([string]$vmName,[string]$integrationServiceId)
Hyper-V\Get-VMIntegrationService -VmName $vmName | ?{$_.Id -match $integrationServiceId} | Hyper-V\Enable-VMIntegrationService
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, integrationServiceId)
	return err
}

func SetNetworkAdapterVlanId(switchName string, vlanId string) error {

	var script = `
param([string]$networkAdapterName,[string]$vlanId)
Hyper-V\Set-VMNetworkAdapterVlan -ManagementOS -VMNetworkAdapterName $networkAdapterName -Access -VlanId $vlanId
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, switchName, vlanId)
	return err
}

func SetVirtualMachineVlanId(vmName string, vlanId string) error {

	var script = `
param([string]$vmName,[string]$vlanId)
Hyper-V\Set-VMNetworkAdapterVlan -VMName $vmName -Access -VlanId $vlanId
`
	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, vlanId)
	return err
}

func ReplaceVirtualMachineNetworkAdapter(vmName string, legacy bool) error {

	var script = `
param([string]$vmName,[string]$legacyString)
$legacy = [System.Boolean]::Parse($legacyString)
$switch = (Get-VMNetworkAdapter -VMName $vmName).SwitchName
Remove-VMNetworkAdapter -VMName $vmName
Add-VMNetworkAdapter -VMName $vmName -SwitchName $switch -Name $vmName -IsLegacy $legacy
`
	legacyString := "False"
	if legacy {
		legacyString = "True"
	}
	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, legacyString)
	return err
}

func GetExternalOnlineVirtualSwitch() (string, error) {

	var script = `
$adapters = Get-NetAdapter -Physical -ErrorAction SilentlyContinue | Where-Object { $_.Status -eq 'Up' } | Sort-Object -Descending -Property Speed
foreach ($adapter in $adapters) {
  $switch = Hyper-V\Get-VMSwitch -SwitchType External | Where-Object { $_.NetAdapterInterfaceDescription -eq $adapter.InterfaceDescription }

  if ($switch -ne $null) {
    $switch.Name
    break
  }
}
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script)
	if err != nil {
		return "", err
	}

	var switchName = strings.TrimSpace(cmdOut)
	return switchName, nil
}

func CreateExternalVirtualSwitch(vmName string, switchName string) error {

	var script = `
param([string]$vmName,[string]$switchName)
$switch = $null
$names = @('ethernet','wi-fi','lan')
$adapters = foreach ($name in $names) {
  Get-NetAdapter -Physical -Name $name -ErrorAction SilentlyContinue | where status -eq 'up'
}

foreach ($adapter in $adapters) {
  $switch = Hyper-V\Get-VMSwitch -SwitchType External | where { $_.NetAdapterInterfaceDescription -eq $adapter.InterfaceDescription }

  if ($switch -eq $null) {
    $switch = Hyper-V\New-VMSwitch -Name $switchName -NetAdapterName $adapter.Name -AllowManagementOS $true -Notes 'Parent OS, VMs, WiFi'
  }

  if ($switch -ne $null) {
    break
  }
}

if($switch -ne $null) {
  Hyper-V\Get-VMNetworkAdapter -VMName $vmName | Hyper-V\Connect-VMNetworkAdapter -VMSwitch $switch
} else {
  Write-Error 'No internet adapters found'
}
`
	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, switchName)
	return err
}

func GetVirtualMachineSwitchName(vmName string) (string, error) {

	var script = `
param([string]$vmName)
(Hyper-V\Get-VMNetworkAdapter -VMName $vmName).SwitchName
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, vmName)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(cmdOut), nil
}

func ConnectVirtualMachineNetworkAdapterToSwitch(vmName string, switchName string) error {

	var script = `
param([string]$vmName,[string]$switchName)
Hyper-V\Get-VMNetworkAdapter -VMName $vmName | Hyper-V\Connect-VMNetworkAdapter -SwitchName $switchName
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, switchName)
	return err
}

func AddVirtualMachineHardDiskDrive(vmName string, vhdRoot string, vhdName string, vhdSizeBytes int64,
	vhdBlockSize int64, controllerType string) error {

	var script = `
param([string]$vmName,[string]$vhdRoot, [string]$vhdName, [string]$vhdSizeInBytes, [string]$vhdBlockSizeInByte, [string]$controllerType)
$vhdPath = Join-Path -Path $vhdRoot -ChildPath $vhdName
Hyper-V\New-VHD -path $vhdPath -SizeBytes $vhdSizeInBytes -BlockSizeBytes $vhdBlockSizeInByte
Hyper-V\Add-VMHardDiskDrive -VMName $vmName -path $vhdPath -controllerType $controllerType
`
	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, vhdRoot, vhdName, strconv.FormatInt(vhdSizeBytes, 10), strconv.FormatInt(vhdBlockSize, 10), controllerType)
	return err
}

func UntagVirtualMachineNetworkAdapterVlan(vmName string, switchName string) error {

	var script = `
param([string]$vmName,[string]$switchName)
Hyper-V\Set-VMNetworkAdapterVlan -VMName $vmName -Untagged
Hyper-V\Set-VMNetworkAdapterVlan -ManagementOS -VMNetworkAdapterName $switchName -Untagged
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, switchName)
	return err
}

func IsRunning(vmName string) (bool, error) {

	var script = `
param([string]$vmName)
$vm = Hyper-V\Get-VM -Name $vmName -ErrorAction SilentlyContinue
$vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Running
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, vmName)

	if err != nil {
		return false, err
	}

	var isRunning = strings.TrimSpace(cmdOut) == "True"
	return isRunning, err
}

func IsOff(vmName string) (bool, error) {

	var script = `
param([string]$vmName)
$vm = Hyper-V\Get-VM -Name $vmName -ErrorAction SilentlyContinue
$vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Off
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, vmName)

	if err != nil {
		return false, err
	}

	var isRunning = strings.TrimSpace(cmdOut) == "True"
	return isRunning, err
}

func Uptime(vmName string) (uint64, error) {

	var script = `
param([string]$vmName)
$vm = Hyper-V\Get-VM -Name $vmName -ErrorAction SilentlyContinue
$vm.Uptime.TotalSeconds
`
	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, vmName)

	if err != nil {
		return 0, err
	}

	uptime, err := strconv.ParseUint(strings.TrimSpace(cmdOut), 10, 64)

	return uptime, err
}

func Mac(vmName string) (string, error) {
	var script = `
param([string]$vmName, [int]$adapterIndex)
try {
  $adapter = Hyper-V\Get-VMNetworkAdapter -VMName $vmName -ErrorAction SilentlyContinue
  $mac = $adapter[$adapterIndex].MacAddress
  if($mac -eq $null) {
    return ""
  }
} catch {
  return ""
}
$mac
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, vmName, "0")

	return cmdOut, err
}

func IpAddress(mac string) (string, error) {
	var script = `
param([string]$mac, [int]$addressIndex)
try {
  $vm = Hyper-V\Get-VM | ?{$_.NetworkAdapters.MacAddress -eq $mac}
  if ($vm.NetworkAdapters.IpAddresses) {
    $ipAddresses = $vm.NetworkAdapters.IPAddresses
    if ($ipAddresses -isnot [array]) {
      $ipAddresses = @($ipAddresses)
    }
    $ip = $ipAddresses[$addressIndex]
  } else {
    $vm_info = Get-CimInstance -ClassName Msvm_ComputerSystem -Namespace root\virtualization\v2 -Filter "ElementName='$($vm.Name)'"
    $ip_details = (Get-CimAssociatedInstance -InputObject $vm_info -ResultClassName Msvm_KvpExchangeComponent).GuestIntrinsicExchangeItems | %{ [xml]$_ } | ?{ $_.SelectSingleNode("/INSTANCE/PROPERTY[@NAME='Name']/VALUE[child::text()='NetworkAddressIPv4']") }

    if ($null -eq $ip_details) {
      return ""
    }

    $ip_addresses = $ip_details.SelectSingleNode("/INSTANCE/PROPERTY[@NAME='Data']/VALUE/child::text()").Value
    $ip = ($ip_addresses -split ";")[0]
  }
} catch {
  return ""
}
$ip
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, mac, "0")

	return cmdOut, err
}

func TurnOff(vmName string) error {

	var script = `
param([string]$vmName)
$vm = Hyper-V\Get-VM -Name $vmName -ErrorAction SilentlyContinue
if ($vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Running) {
  Hyper-V\Stop-VM -Name $vmName -TurnOff -Force -Confirm:$false
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func ShutDown(vmName string) error {

	var script = `
param([string]$vmName)
$vm = Hyper-V\Get-VM -Name $vmName -ErrorAction SilentlyContinue
if ($vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Running) {
  Hyper-V\Stop-VM -Name $vmName -Force -Confirm:$false
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func TypeScanCodes(vmName string, scanCodes string) error {
	if len(scanCodes) == 0 {
		return nil
	}

	var script = `
param([string]$vmName, [string]$scanCodes)
	#Requires -Version 3

	function Hyper-V\Get-VMConsole
	{
	    [CmdletBinding()]
	    param (
	        [Parameter(Mandatory)]
	        [string] $VMName
	    )

	    $ErrorActionPreference = "Stop"

	    $vm = Get-CimInstance -Namespace "root\virtualization\v2" -ClassName Msvm_ComputerSystem -ErrorAction Ignore -Verbose:$false | where ElementName -eq $VMName | select -first 1
	    if ($vm -eq $null){
	        Write-Error ("VirtualMachine({0}) is not found!" -f $VMName)
	    }

	    $vmKeyboard = $vm | Get-CimAssociatedInstance -ResultClassName "Msvm_Keyboard" -ErrorAction Ignore -Verbose:$false

		if ($vmKeyboard -eq $null) {
			$vmKeyboard = Get-CimInstance -Namespace "root\virtualization\v2" -ClassName Msvm_Keyboard -ErrorAction Ignore -Verbose:$false | where SystemName -eq $vm.Name | select -first 1
		}

		if ($vmKeyboard -eq $null) {
			$vmKeyboard = Get-CimInstance -Namespace "root\virtualization" -ClassName Msvm_Keyboard -ErrorAction Ignore -Verbose:$false | where SystemName -eq $vm.Name | select -first 1
		}

	    if ($vmKeyboard -eq $null){
	        Write-Error ("VirtualMachine({0}) keyboard class is not found!" -f $VMName)
	    }

	    #TODO: It may be better using New-Module -AsCustomObject to return console object?

	    #Console object to return
	    $console = [pscustomobject] @{
	        Msvm_ComputerSystem = $vm
	        Msvm_Keyboard = $vmKeyboard
	    }

	    #Need to import assembly to use System.Windows.Input.Key
	    Add-Type -AssemblyName WindowsBase

	    #region Add Console Members
	    $console | Add-Member -MemberType ScriptMethod -Name TypeText -Value {
	        [OutputType([bool])]
	        param (
	            [ValidateNotNullOrEmpty()]
	            [Parameter(Mandatory)]
	            [string] $AsciiText
	        )
	        $result = $this.Msvm_Keyboard | Invoke-CimMethod -MethodName "TypeText" -Arguments @{ asciiText = $AsciiText }
	        return (0 -eq $result.ReturnValue)
	    }

	    #Define method:TypeCtrlAltDel
	    $console | Add-Member -MemberType ScriptMethod -Name TypeCtrlAltDel -Value {
	        $result = $this.Msvm_Keyboard | Invoke-CimMethod -MethodName "TypeCtrlAltDel"
	        return (0 -eq $result.ReturnValue)
	    }

	    #Define method:TypeKey
	    $console | Add-Member -MemberType ScriptMethod -Name TypeKey -Value {
	        [OutputType([bool])]
	        param (
	            [Parameter(Mandatory)]
	            [Windows.Input.Key] $Key,
	            [Windows.Input.ModifierKeys] $ModifierKey = [Windows.Input.ModifierKeys]::None
	        )

	        $keyCode = [Windows.Input.KeyInterop]::VirtualKeyFromKey($Key)

	        switch ($ModifierKey)
	        {
	            ([Windows.Input.ModifierKeys]::Control){ $modifierKeyCode = [Windows.Input.KeyInterop]::VirtualKeyFromKey([Windows.Input.Key]::LeftCtrl)}
	            ([Windows.Input.ModifierKeys]::Alt){ $modifierKeyCode = [Windows.Input.KeyInterop]::VirtualKeyFromKey([Windows.Input.Key]::LeftAlt)}
	            ([Windows.Input.ModifierKeys]::Shift){ $modifierKeyCode = [Windows.Input.KeyInterop]::VirtualKeyFromKey([Windows.Input.Key]::LeftShift)}
	            ([Windows.Input.ModifierKeys]::Windows){ $modifierKeyCode = [Windows.Input.KeyInterop]::VirtualKeyFromKey([Windows.Input.Key]::LWin)}
	        }

	        if ($ModifierKey -eq [Windows.Input.ModifierKeys]::None)
	        {
	            $result = $this.Msvm_Keyboard | Invoke-CimMethod -MethodName "TypeKey" -Arguments @{ keyCode = $keyCode }
	        }
	        else
	        {
	            $this.Msvm_Keyboard | Invoke-CimMethod -MethodName "PressKey" -Arguments @{ keyCode = $modifierKeyCode }
	            $result = $this.Msvm_Keyboard | Invoke-CimMethod -MethodName "TypeKey" -Arguments @{ keyCode = $keyCode }
	            $this.Msvm_Keyboard | Invoke-CimMethod -MethodName "ReleaseKey" -Arguments @{ keyCode = $modifierKeyCode }
	        }
	        $result = return (0 -eq $result.ReturnValue)
	    }

	    #Define method:Scancodes
	    $console | Add-Member -MemberType ScriptMethod -Name TypeScancodes -Value {
	        [OutputType([bool])]
	        param (
	            [Parameter(Mandatory)]
	            [byte[]] $ScanCodes
	        )
	        $result = $this.Msvm_Keyboard | Invoke-CimMethod -MethodName "TypeScancodes" -Arguments @{ ScanCodes = $ScanCodes }
	        return (0 -eq $result.ReturnValue)
	    }

	    #Define method:ExecCommand
	    $console | Add-Member -MemberType ScriptMethod -Name ExecCommand -Value {
	        param (
	            [Parameter(Mandatory)]
	            [string] $Command
	        )
	        if ([String]::IsNullOrEmpty($Command)){
	            return
	        }

	        $console.TypeText($Command) > $null
	        $console.TypeKey([Windows.Input.Key]::Enter) > $null
	        #sleep -Milliseconds 100
	    }

	    #Define method:Dispose
	    $console | Add-Member -MemberType ScriptMethod -Name Dispose -Value {
	        $this.Msvm_ComputerSystem.Dispose()
	        $this.Msvm_Keyboard.Dispose()
	    }


	    #endregion

	    return $console
	}

	$vmConsole = Hyper-V\Get-VMConsole -VMName $vmName
	$scanCodesToSend = ''
	$scanCodes.Split(' ') | %{
		$scanCode = $_

		if ($scanCode.StartsWith('wait')){
			$timeToWait = $scanCode.Substring(4)
			if (!$timeToWait){
				$timeToWait = "1"
			}

			if ($scanCodesToSend){
				$scanCodesToSendByteArray = [byte[]]@($scanCodesToSend.Split(' ') | %{"0x$_"})

                $scanCodesToSendByteArray | %{
				    $vmConsole.TypeScancodes($_)
                }
			}

			write-host "Special code <wait> found, will sleep $timeToWait second(s) at this point."
			Start-Sleep -s $timeToWait

			$scanCodesToSend = ''
		} else {
			if ($scanCodesToSend){
				write-host "Sending special code '$scanCodesToSend' '$scanCode'"
				$scanCodesToSend = "$scanCodesToSend $scanCode"
			} else {
				write-host "Sending char '$scanCode'"
				$scanCodesToSend = "$scanCode"
			}
		}
	}
	if ($scanCodesToSend){
		$scanCodesToSendByteArray = [byte[]]@($scanCodesToSend.Split(' ') | %{"0x$_"})

        $scanCodesToSendByteArray | %{
			$vmConsole.TypeScancodes($_)
        }
	}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, scanCodes)
	return err
}

func ConnectVirtualMachine(vmName string) (context.CancelFunc, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "vmconnect.exe", "localhost", vmName)
	err := cmd.Start()
	if err != nil {
		// Failed to start so cancel function not required
		cancel = nil
	}
	return cancel, err
}

func DisconnectVirtualMachine(cancel context.CancelFunc) {
	cancel()
}
