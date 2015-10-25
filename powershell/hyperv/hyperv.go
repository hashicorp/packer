package hyperv

import (
	"github.com/mitchellh/packer/powershell"
	"strings"
)

func GetHostAdapterIpAddressForSwitch(switchName string) (string, error) {
	var script = `
param([string]$switchName, [int]$addressIndex)

$HostVMAdapter = Get-VMNetworkAdapter -ManagementOS -SwitchName $switchName
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
  $adapter = Get-VMNetworkAdapter -VMName $vmName -ErrorAction SilentlyContinue
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

func MountDvdDrive(vmName string, path string) error {

	var script = `
param([string]$vmName,[string]$path)
Set-VMDvdDrive -VMName $vmName -Path $path
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, path)
	return err
}

func UnmountDvdDrive(vmName string) error {
	var script = `
param([string]$vmName)
Get-VMDvdDrive -VMName $vmName | Set-VMDvdDrive  -Path $null
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func MountFloppyDrive(vmName string, path string) error {
	var script = `
param([string]$vmName, [string]$path)
Set-VMFloppyDiskDrive -VMName $vmName -Path $path
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, path)
	return err
}

func UnmountFloppyDrive(vmName string) error {

	var script = `
param([string]$vmName)
Set-VMFloppyDiskDrive -VMName $vmName -Path $null
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func CreateVirtualMachine(vmName string, path string, ram string, diskSize string, switchName string, generation string) error {

	if generation == "2" {
		var script = `
param([string]$vmName, [string]$path, [long]$memoryStartupBytes, [long]$newVHDSizeBytes, [string]$switchName, [int]$generation)
$vhdx = $vmName + '.vhdx'
$vhdPath = Join-Path -Path $path -ChildPath $vhdx
New-VM -Name $vmName -Path $path -MemoryStartupBytes $memoryStartupBytes -NewVHDPath $vhdPath -NewVHDSizeBytes $newVHDSizeBytes -SwitchName $switchName -Generation $generation
`
		var ps powershell.PowerShellCmd
		err := ps.Run(script, vmName, path, ram, diskSize, switchName, generation)
		return err
	} else {
		var script = `
param([string]$vmName, [string]$path, [long]$memoryStartupBytes, [long]$newVHDSizeBytes, [string]$switchName)
$vhdx = $vmName + '.vhdx'
$vhdPath = Join-Path -Path $path -ChildPath $vhdx
New-VM -Name $vmName -Path $path -MemoryStartupBytes $memoryStartupBytes -NewVHDPath $vhdPath -NewVHDSizeBytes $newVHDSizeBytes -SwitchName $switchName
`
		var ps powershell.PowerShellCmd
		err := ps.Run(script, vmName, path, ram, diskSize, switchName)
		return err
	}
}

func SetVirtualMachineCpu(vmName string, cpu string) error {

	var script = `
param([string]$vmName, [int]$cpu)
Set-VMProcessor -VMName $vmName -Count $cpu
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, cpu)
	return err
}

func SetSecureBoot(vmName string, enable bool) error {
	var script = `
param([string]$vmName, $enableSecureBoot)
Set-VMFirmware -VMName $vmName -EnableSecureBoot $enableSecureBoot
`

	var ps powershell.PowerShellCmd

	enableSecureBoot := "Off"
	if enable {
		enableSecureBoot = "On"
	}

	err := ps.Run(script, vmName, enableSecureBoot)
	return err
}

func DeleteVirtualMachine(vmName string) error {

	var script = `
param([string]$vmName)

$vm = Get-VM -Name $vmName
if (($vm.State -ne [Microsoft.HyperV.PowerShell.VMState]::Off) -and ($vm.State -ne [Microsoft.HyperV.PowerShell.VMState]::OffCritical)) {
    Stop-VM -VM $vm -TurnOff -Force -Confirm:$false
}

Remove-VM -Name $vmName -Force -Confirm:$false
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func ExportVirtualMachine(vmName string, path string) error {

	var script = `
param([string]$vmName, [string]$path)
Export-VM -Name $vmName -Path $path
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
$switches = Get-VMSwitch -Name $switchName -ErrorAction SilentlyContinue
if ($switches.Count -eq 0) {
  New-VMSwitch -Name $switchName -SwitchType $switchType
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
$switch = Get-VMSwitch -Name $switchName -ErrorAction SilentlyContinue
if ($switch -ne $null) {
    $switch | Remove-VMSwitch -Force -Confirm:$false
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, switchName)
	return err
}

func StartVirtualMachine(vmName string) error {

	var script = `
param([string]$vmName)
$vm = Get-VM -Name $vmName -ErrorAction SilentlyContinue
if ($vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Off) {
  Start-VM -Name $vmName -Confirm:$false
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func RestartVirtualMachine(vmName string) error {

	var script = `
param([string]$vmName)
Restart-VM $vmName -Force -Confirm:$false
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func StopVirtualMachine(vmName string) error {

	var script = `
param([string]$vmName)
$vm = Get-VM -Name $vmName
if ($vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Running) {
    Stop-VM -VM $vm -Confirm:$false
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func EnableVirtualMachineIntegrationService(vmName string, integrationServiceName string) error {

	var script = `
param([string]$vmName,[string]$integrationServiceName)
Enable-VMIntegrationService -VMName $vmName -Name $integrationServiceName
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, integrationServiceName)
	return err
}

func SetNetworkAdapterVlanId(switchName string, vlanId string) error {

	var script = `
param([string]$networkAdapterName,[string]$vlanId)
Set-VMNetworkAdapterVlan -ManagementOS -VMNetworkAdapterName $networkAdapterName -Access -VlanId $vlanId
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, switchName, vlanId)
	return err
}

func SetVirtualMachineVlanId(vmName string, vlanId string) error {

	var script = `
param([string]$vmName,[string]$vlanId)
Set-VMNetworkAdapterVlan -VMName $vmName -Access -VlanId $vlanId
`
	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, vlanId)
	return err
}

func GetExternalOnlineVirtualSwitch() (string, error) {

	var script = `
$adapters = Get-NetAdapter -Physical -ErrorAction SilentlyContinue | Where-Object { $_.Status -eq 'Up' } | Sort-Object -Descending -Property Speed
foreach ($adapter in $adapters) { 
  $switch = Get-VMSwitch -SwitchType External | Where-Object { $_.NetAdapterInterfaceDescription -eq $adapter.InterfaceDescription }

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
  $switch = Get-VMSwitch -SwitchType External | where { $_.NetAdapterInterfaceDescription -eq $adapter.InterfaceDescription }

  if ($switch -eq $null) { 
    $switch = New-VMSwitch -Name $switchName -NetAdapterName $adapter.Name -AllowManagementOS $true -Notes 'Parent OS, VMs, WiFi'
  }

  if ($switch -ne $null) {
    break
  }
}

if($switch -ne $null) { 
  Get-VMNetworkAdapter -VMName $vmName | Connect-VMNetworkAdapter -VMSwitch $switch 
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
(Get-VMNetworkAdapter -VMName $vmName).SwitchName
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
Get-VMNetworkAdapter -VMName $vmName | Connect-VMNetworkAdapter -SwitchName $switchName
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, switchName)
	return err
}

func UntagVirtualMachineNetworkAdapterVlan(vmName string, switchName string) error {

	var script = `
param([string]$vmName,[string]$switchName)
Set-VMNetworkAdapterVlan -VMName $vmName -Untagged
Set-VMNetworkAdapterVlan -ManagementOS -VMNetworkAdapterName $switchName -Untagged
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName, switchName)
	return err
}

func IsRunning(vmName string) (bool, error) {

	var script = `
param([string]$vmName)
$vm = Get-VM -Name $vmName -ErrorAction SilentlyContinue
$vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Running
`

	var ps powershell.PowerShellCmd
	cmdOut, err := ps.Output(script, vmName)
	var isRunning = strings.TrimSpace(cmdOut) == "True"
	return isRunning, err
}

func Mac(vmName string) (string, error) {
	var script = `
param([string]$vmName, [int]$adapterIndex)
try {
  $adapter = Get-VMNetworkAdapter -VMName $vmName -ErrorAction SilentlyContinue
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
  $ip = Get-Vm | %{$_.NetworkAdapters} | ?{$_.MacAddress -eq $mac} | %{$_.IpAddresses[$addressIndex]}
	
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
$vm = Get-VM -Name $vmName -ErrorAction SilentlyContinue
if ($vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Running) {
  Stop-VM -Name $vmName -TurnOff -Confirm:$false -Force
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func ShutDown(vmName string) error {

	var script = `
param([string]$vmName)
$vm = Get-VM -Name $vmName -ErrorAction SilentlyContinue
if ($vm.State -eq [Microsoft.HyperV.PowerShell.VMState]::Running) {
  Stop-VM -Name $vmName -Confirm:$false -Force
}
`

	var ps powershell.PowerShellCmd
	err := ps.Run(script, vmName)
	return err
}

func TypeScanCodes(vmName string, scanCodes string) error {
	var script = `
param([string]$vmName, [string]$scanCodes)
	#Requires -Version 3
	#Requires -RunAsAdministrator
	
	function Get-VMConsole
	{
	    [CmdletBinding()]
	    param (
	        [Parameter(Mandatory)]
	        [string] $VMName
	    )
	
	    $ErrorActionPreference = "Stop"
	    
	    $vm = Get-CimInstance -ComputerName localhost -Namespace "root\virtualization\v2" -ClassName Msvm_ComputerSystem -ErrorAction Ignore -Verbose:$false | where ElementName -eq $VMName | select -first 1
	    if ($vm -eq $null){
	        Write-Error ("VirtualMachine({0}) is not found!" -f $VMName)
	    }
	
	    $vmKeyboard = $vm | Get-CimAssociatedInstance -ResultClassName "Msvm_Keyboard" -ErrorAction Ignore -Verbose:$false
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
	
	$vmConsole = Get-VMConsole -VMName $vmName
	$scanCodesToSend = ''
	$scanCodes.Split(' ') | %{
		$scanCode = $_
			
		if ($scanCode.StartsWith('wait')){
			$timeToWait = $scanCode.Substring(4)
			if (!$timeToWait){
				$timeToWait = "10"
			}
			
			Start-Sleep -s $timeToWait
			
			if ($scanCodesToSend){
				$scanCodesToSendByteArray = [byte[]]@($scanCodesToSend.Split(' ') | %{"0x$_"})
				
                $scanCodesToSendByteArray | %{
				    $vmConsole.TypeScancodes($_)
                }
			}
			
			$scanCodesToSend = ''
		} else {
			if ($scanCodesToSend){
				$scanCodesToSend = "$scanCodesToSend $scanCode"
			} else {
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
