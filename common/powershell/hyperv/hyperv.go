package hyperv

import (
	"errors"
	"strconv"
	"strings"

	"github.com/hashicorp/packer/common/powershell"
)

func GetHostAdapterIpAddressForSwitch(switchName string) (string, error) {
	var script = `
param([string]$switchName, [int]$addressIndex)

$HostVMAdapter = Hyper-V\Get-VMNetworkAdapter -ManagementOS -SwitchName $switchName
if ($HostVMAdapter){
    $HostNetAdapter = Get-NetAdapter | ?{ $_.DeviceID -eq $HostVMAdapter.DeviceId }
    if ($HostNetAdapter){
        $HostNetAdapterConfiguration =  @(get-wmiobject win32_networkadapterconfiguration -filter "IPEnabled = 'TRUE' AND InterfaceIndex=$($HostNetAdapter.ifIndex)")
        if ($HostNetAdapterConfiguration){
            return @($HostNetAdapterConfiguration.IpAddress)[$addressIndex]
        }
    }
}
return $false
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
  $ip = $adapter.IPAddresses[$addressIndex]
  if($ip -eq $null) {
    return $false
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
	err := ps.Run(script, vmName, path, strconv.FormatInt(int64(controllerNumber), 10), strconv.FormatInt(int64(controllerLocation), 10))
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
	err := ps.Run(script, vmName, strconv.FormatInt(int64(controllerNumber), 10), strconv.FormatInt(int64(controllerLocation), 10))
	return err
}

func SetBootDvdDrive(vmName string, controllerNumber uint, controllerLocation uint, generation uint) error {

	if generation < 2 {
		script := `
param([string]$vmName)
Hyper-V\Set-VMBios -VMName $vmName -StartupOrder @("CD", "IDE","LegacyNetworkAdapter","Floppy")
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
		err := ps.Run(script, vmName, strconv.FormatInt(int64(controllerNumber), 10), strconv.FormatInt(int64(controllerLocation), 10))
		return err
	}
}

func DeleteDvdDrive(vmName string, controllerNumber uint, controllerLocation uint) error {
	var script = `
param([string]$vmName,[int]$controllerNumber,[int]$controllerLocation)
$vmDvdDrive = Hyper-V\Get-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation
if (!$vmDvdDrive) {throw 'unable to find dvd drive'}
Hyper-V\Remove-VMDvdDrive -VMName $vmName -ControllerNumber $controllerNumber -ControllerLocation $controllerLocation
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, strconv.FormatInt(int64(controllerNumber), 10), strconv.FormatInt(int64(controllerLocation), 10))
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

func CreateVirtualMachine(vmName string, path string, harddrivePath string, vhdRoot string, ram int64, diskSize int64, switchName string, generation uint, diffDisks bool) error {

	if generation == 2 {
		var script = `
param([string]$vmName, [string]$path, [string]$harddrivePath, [string]$vhdRoot, [long]$memoryStartupBytes, [long]$newVHDSizeBytes, [string]$switchName, [int]$generation, [string]$diffDisks)
$vhdx = $vmName + '.vhdx'
$vhdPath = Join-Path -Path $vhdRoot -ChildPath $vhdx
if ($harddrivePath){
	if($diffDisks -eq "true"){
		New-VHD -Path $vhdPath -ParentPath $harddrivePath -Differencing
	} else {
		Copy-Item -Path $harddrivePath -Destination $vhdPath
	}
	Hyper-V\New-VM -Name $vmName -Path $path -MemoryStartupBytes $memoryStartupBytes -VHDPath $vhdPath -SwitchName $switchName -Generation $generation
} else {
	Hyper-V\New-VM -Name $vmName -Path $path -MemoryStartupBytes $memoryStartupBytes -NewVHDPath $vhdPath -NewVHDSizeBytes $newVHDSizeBytes -SwitchName $switchName -Generation $generation
}
`
		var ps powershell.PowerShellCmd
		if err := ps.Run(script, vmName, path, harddrivePath, vhdRoot, strconv.FormatInt(ram, 10), strconv.FormatInt(diskSize, 10), switchName, strconv.FormatInt(int64(generation), 10), strconv.FormatBool(diffDisks)); err != nil {
			return err
		}

		return DisableAutomaticCheckpoints(vmName)
	} else {
		var script = `
param([string]$vmName, [string]$path, [string]$harddrivePath, [string]$vhdRoot, [long]$memoryStartupBytes, [long]$newVHDSizeBytes, [string]$switchName, [string]$diffDisks)
$vhdx = $vmName + '.vhdx'
$vhdPath = Join-Path -Path $vhdRoot -ChildPath $vhdx
if ($harddrivePath){
	if($diffDisks -eq "true"){
		New-VHD -Path $vhdPath -ParentPath $harddrivePath -Differencing
	}
	else{
		Copy-Item -Path $harddrivePath -Destination $vhdPath
	}
	Hyper-V\New-VM -Name $vmName -Path $path -MemoryStartupBytes $memoryStartupBytes -VHDPath $vhdPath -SwitchName $switchName
} else {
	Hyper-V\New-VM -Name $vmName -Path $path -MemoryStartupBytes $memoryStartupBytes -NewVHDPath $vhdPath -NewVHDSizeBytes $newVHDSizeBytes -SwitchName $switchName
}
`
		var ps powershell.PowerShellCmd
		if err := ps.Run(script, vmName, path, harddrivePath, vhdRoot, strconv.FormatInt(ram, 10), strconv.FormatInt(diskSize, 10), switchName, strconv.FormatBool(diffDisks)); err != nil {
			return err
		}

		if err := DisableAutomaticCheckpoints(vmName); err != nil {
			return err
		}

		return DeleteAllDvdDrives(vmName)
	}
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

func ExportVmxcVirtualMachine(exportPath string, vmName string, snapshotName string, allSnapshots bool) error {
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

func CopyVmxcVirtualMachine(exportPath string, cloneFromVmxcPath string) error {
	var script = `
param([string]$exportPath, [string]$cloneFromVmxcPath)
if (!(Test-Path $cloneFromVmxcPath)){
	throw "Clone from vmxc directory: $cloneFromVmxcPath does not exist!"
}
	
if (!(Test-Path $exportPath)){
	New-Item -ItemType Directory -Force -Path $exportPath
}
$cloneFromVmxcPath = Join-Path $cloneFromVmxcPath '\*'
Copy-Item $cloneFromVmxcPath $exportPath -Recurse -Force
	`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, exportPath, cloneFromVmxcPath)

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

func ImportVmxcVirtualMachine(importPath string, vmName string, harddrivePath string, ram int64, switchName string) error {
	var script = `
param([string]$importPath, [string]$vmName, [string]$harddrivePath, [long]$memoryStartupBytes, [string]$switchName)

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
	err := ps.Run(script, importPath, vmName, harddrivePath, strconv.FormatInt(ram, 10), switchName)

	return err
}

func CloneVirtualMachine(cloneFromVmxcPath string, cloneFromVmName string, cloneFromSnapshotName string, cloneAllSnapshots bool, vmName string, path string, harddrivePath string, ram int64, switchName string) error {
	if cloneFromVmName != "" {
		if err := ExportVmxcVirtualMachine(path, cloneFromVmName, cloneFromSnapshotName, cloneAllSnapshots); err != nil {
			return err
		}
	}

	if cloneFromVmxcPath != "" {
		if err := CopyVmxcVirtualMachine(path, cloneFromVmxcPath); err != nil {
			return err
		}
	}

	if err := ImportVmxcVirtualMachine(path, vmName, harddrivePath, ram, switchName); err != nil {
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

func SetVirtualMachineSecureBoot(vmName string, enableSecureBoot bool) error {
	var script = `
param([string]$vmName, $enableSecureBoot)
Hyper-V\Set-VMFirmware -VMName $vmName -EnableSecureBoot $enableSecureBoot
`

	var ps powershell.PowerShellCmd

	enableSecureBootString := "Off"
	if enableSecureBoot {
		enableSecureBootString = "On"
	}

	err := ps.Run(script, vmName, enableSecureBootString)
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

func CompactDisks(expPath string, vhdDir string) error {
	var script = `
param([string]$srcPath, [string]$vhdDirName)
Get-ChildItem "$srcPath/$vhdDirName" -Filter *.vhd* | %{
    Optimize-VHD -Path $_.FullName -Mode Full
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, expPath, vhdDir)
	return err
}

func CopyExportedVirtualMachine(expPath string, outputPath string, vhdDir string, vmDir string) error {

	var script = `
param([string]$srcPath, [string]$dstPath, [string]$vhdDirName, [string]$vmDir)
Move-Item -Path $srcPath/*.* -Destination $dstPath
Move-Item -Path $srcPath/$vhdDirName -Destination $dstPath
Move-Item -Path $srcPath/$vmDir -Destination $dstPath
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, expPath, outputPath, vhdDir, vmDir)
	return err
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

func AddVirtualMachineHardDiskDrive(vmName string, vhdRoot string, vhdName string, vhdSizeBytes int64, controllerType string) error {

	var script = `
param([string]$vmName,[string]$vhdRoot, [string]$vhdName, [string]$vhdSizeInBytes, [string]$controllerType)
$vhdPath = Join-Path -Path $vhdRoot -ChildPath $vhdName
New-VHD $vhdPath -SizeBytes $vhdSizeInBytes
Hyper-V\Add-VMHardDiskDrive -VMName $vmName -path $vhdPath -controllerType $controllerType
`
	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, vhdRoot, vhdName, strconv.FormatInt(vhdSizeBytes, 10), controllerType)
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
  $ip = Hyper-V\Get-Vm | %{$_.NetworkAdapters} | ?{$_.MacAddress -eq $mac} | %{$_.IpAddresses[$addressIndex]}

  if($ip -eq $null) {
    return ""
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
