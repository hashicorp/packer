---
description: |-
  The HyperV Packer builder is able to create HyperV virtual machines and export them.
layout: "docs"
page_title: "HyperV Builder (from an ISO)"
---

# HyperV Builder (from an ISO)

Type: `hyperv-iso`

The HyperV Packer builder is able to create [HyperV](https://www.microsoft.com/en-us/server-cloud/solutions/virtualization.aspx)
virtual machines and export them, starting from an ISO image.

The builder builds a virtual machine by creating a new virtual machine
from scratch, booting it, installing an OS, provisioning software within
the OS, then shutting it down. The result of the HyperV builder is a directory
containing all the files necessary to run the virtual machine portably.

## Basic Example

Here is a basic example. This example is not functional. It will start the
OS installer but then fail because we don't provide the preseed file for
Ubuntu to self-install. Still, the example serves to show the basic configuration:

```javascript
{
  "type": "hyperv-iso",
  "guest_os_type": "Ubuntu_64",
  "iso_url": "http://releases.ubuntu.com/12.04/ubuntu-12.04.5-server-amd64.iso",
  "iso_checksum": "769474248a3897f4865817446f9a4a53",
  "iso_checksum_type": "md5",
  "ssh_username": "packer",
  "ssh_password": "packer",
  "shutdown_command": "echo 'packer' | sudo -S shutdown -P now"
}
```

It is important to add a `shutdown_command`. By default Packer halts the
virtual machine and the file system may not be sync'd. Thus, changes made in a
provisioner might not be saved.

## Configuration Reference

There are many configuration options available for the HyperV builder.
They are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html)
can be configured for this builder.

### Required:

-   `iso_checksum` (string) - The checksum for the OS ISO file. Because ISO
    files are so large, this is required and Packer will verify it prior
    to booting a virtual machine with the ISO attached. The type of the
    checksum is specified with `iso_checksum_type`, documented below.

-   `iso_checksum_type` (string) - The type of the checksum specified in
    `iso_checksum`. Valid values are "none", "md5", "sha1", "sha256", or
    "sha512" currently. While "none" will skip checksumming, this is not
    recommended since ISO files are generally large and corruption does happen
    from time to time.

-   `iso_url` (string) - A URL to the ISO containing the installation image.
    This URL can be either an HTTP URL or a file URL (or path to a file).
    If this is an HTTP URL, Packer will download it and cache it between
    runs.

### Optional:

-   `boot_command` (array of strings) - This is an array of commands to type
    when the virtual machine is first booted. The goal of these commands should
    be to type just enough to initialize the operating system installer. Special
    keys can be typed as well, and are covered in the section below on the boot
    command. If this is not specified, it is assumed the installer will start
    itself.

-   `boot_wait` (string) - The time to wait after booting the initial virtual
    machine before typing the `boot_command`. The value of this should be
    a duration. Examples are "5s" and "1m30s" which will cause Packer to wait
    five seconds and one minute 30 seconds, respectively. If this isn't specified,
    the default is 10 seconds.

-   `cpu` (integer) - The number of cpus the virtual machine should use. If this isn't specified,
    the default is 1 cpu.

-   `disk_size` (integer) - The size, in megabytes, of the hard disk to create
    for the VM. By default, this is 40000 (about 40 GB).

-   `enable_mac_spoofing` (bool) - If true enable mac spoofing for virtual machine.
    This defaults to false.

-   `enable_dynamic_memory` (bool) - If true enable dynamic memory for virtual machine.
    This defaults to false.

-   `enable_secure_boot` (bool) - If true enable secure boot for virtual machine.
    This defaults to false.

-   `enable_virtualization_extensions` (bool) - If true enable virtualization extensions for virtual machine.
    This defaults to false. For nested virtualization you need to enable mac spoofing, disable dynamic memory
    and have at least 4GB of RAM for virtual machine.

-   `floppy_files` (array of strings) - A list of files to place onto a floppy
    disk that is attached when the VM is booted. This is most useful
    for unattended Windows installs, which look for an `Autounattend.xml` file
    on removable media. By default, no floppy will be attached. All files
    listed in this setting get placed into the root directory of the floppy
    and the floppy is attached as the first floppy device. Currently, no
    support exists for creating sub-directories on the floppy. Wildcard
    characters (*, ?, and []) are allowed. Directory names are also allowed,
    which will add all the files found in the directory to the floppy.

-   `generation` (integer) - The HyperV generation for the virtual machine. By
    default, this is 1. Generation 2 HyperV virtual machines do not support
    floppy drives. In this scenario use `secondary_iso_images` instead. Hard
    drives and dvd drives will also be scsi and not ide. 

-   `http_directory` (string) - Path to a directory to serve using an HTTP
    server. The files in this directory will be available over HTTP that will
    be requestable from the virtual machine. This is useful for hosting
    kickstart files and so on. By default this is "", which means no HTTP
    server will be started. The address and port of the HTTP server will be
    available as variables in `boot_command`. This is covered in more detail
    below.

-   `http_port_min` and `http_port_max` (integer) - These are the minimum and
    maximum port to use for the HTTP server started to serve the `http_directory`.
    Because Packer often runs in parallel, Packer will choose a randomly available
    port in this range to run the HTTP server. If you want to force the HTTP
    server to be on one port, make this minimum and maximum port the same.
    By default the values are 8000 and 9000, respectively.

-   `ip_address_timeout` (string) - The time to wait after creating the initial virtual
    machine and waiting for an ip address before assuming there is an error in the process. 
    The value of this should be a duration. Examples are "5s" and "1m30s" which will cause Packer to wait
    five seconds and one minute 30 seconds, respectively. If this isn't specified,
    the default is 10 seconds.

-   `iso_urls` (array of strings) - Multiple URLs for the ISO to download.
    Packer will try these in order. If anything goes wrong attempting to download
    or while downloading a single URL, it will move on to the next. All URLs
    must point to the same file (same checksum). By default this is empty
    and `iso_url` is used. Only one of `iso_url` or `iso_urls` can be specified.

-   `output_directory` (string) - This is the path to the directory where the
    resulting virtual machine will be created. This may be relative or absolute.
    If relative, the path is relative to the working directory when `packer`
    is executed. This directory must not exist or be empty prior to running the builder.
    By default this is "output-BUILDNAME" where "BUILDNAME" is the name
    of the build.

*   `secondary_iso_images` (array of strings) - A list of iso paths to attached to a 
    VM when it is booted. This is most useful for unattended Windows installs, which 
    look for an `Autounattend.xml` file on removable media. By default, no 
    secondary iso will be attached. 

-   `shutdown_command` (string) - The command to use to gracefully shut down the machine once all
    the provisioning is done. By default this is an empty string, which tells Packer to just
    forcefully shut down the machine unless a shutdown command takes place inside script so this may
    safely be omitted. If one or more scripts require a reboot it is suggested to leave this blank
    since reboots may fail and specify the final shutdown command in your last script.

-   `shutdown_timeout` (string) - The amount of time to wait after executing
    the `shutdown_command` for the virtual machine to actually shut down.
    If it doesn't shut down in this time, it is an error. By default, the timeout
    is "5m", or five minutes.

-   `skip_compaction` (bool) - If true skip compacting the hard disk for virtual machine when
    exporting. This defaults to false.

-   `switch_name` (string) - The name of the switch to connect the virtual machine to. Be defaulting 
    this to an empty string, Packer will try to determine the switch to use by looking for
    external switch that is up and running.

-   `switch_vlan_id` (string) - This is the vlan of the virtual switch's network card. 
    By default none is set. If none is set then a vlan is not set on the switch's network card.
    If this value is set it should match the vlan specified in by `vlan_id`.

-   `vlan_id` (string) - This is the vlan of the virtual machine's network card for the new virtual
    machine. By default none is set. If none is set then vlans are not set on the virtual machine's 
    network card.

-   `vm_name` (string) - This is the name of the virtual machine for the new virtual
    machine, without the file extension. By default this is "packer-BUILDNAME",
    where "BUILDNAME" is the name of the build.

## Boot Command

The `boot_command` configuration is very important: it specifies the keys
to type when the virtual machine is first booted in order to start the
OS installer. This command is typed after `boot_wait`, which gives the
virtual machine some time to actually load the ISO.

As documented above, the `boot_command` is an array of strings. The
strings are all typed in sequence. It is an array only to improve readability
within the template.

The boot command is "typed" character for character over the virtual keyboard
to the machine, simulating a human actually typing the keyboard. There are
a set of special keys available. If these are in your boot command, they
will be replaced by the proper key:

-   `<bs>` - Backspace

-   `<del>` - Delete

-   `<enter>` and `<return>` - Simulates an actual "enter" or "return" keypress.

-   `<esc>` - Simulates pressing the escape key.

-   `<tab>` - Simulates pressing the tab key.

-   `<f1>` - `<f12>` - Simulates pressing a function key.

-   `<up>` `<down>` `<left>` `<right>` - Simulates pressing an arrow key.

-   `<spacebar>` - Simulates pressing the spacebar.

-   `<insert>` - Simulates pressing the insert key.

-   `<home>` `<end>` - Simulates pressing the home and end keys.

-   `<pageUp>` `<pageDown>` - Simulates pressing the page up and page down keys.

-   `<leftAlt>` `<rightAlt>`  - Simulates pressing the alt key.

-   `<leftCtrl>` `<rightCtrl>` - Simulates pressing the ctrl key.

-   `<leftShift>` `<rightShift>` - Simulates pressing the shift key.

-   `<leftAltOn>` `<rightAltOn>`  - Simulates pressing and holding the alt key.

-   `<leftCtrlOn>` `<rightCtrlOn>` - Simulates pressing and holding the ctrl key. 

-   `<leftShiftOn>` `<rightShiftOn>` - Simulates pressing and holding the shift key.

-   `<leftAltOff>` `<rightAltOff>`  - Simulates releasing a held alt key.

-   `<leftCtrlOff>` `<rightCtrlOff>` - Simulates releasing a held ctrl key.

-   `<leftShiftOff>` `<rightShiftOff>` - Simulates releasing a held shift key.

-   `<wait>` `<wait5>` `<wait10>` - Adds a 1, 5 or 10 second pause before
    sending any additional keys. This is useful if you have to generally wait
    for the UI to update before typing more.

When using modifier keys `ctrl`, `alt`, `shift` ensure that you release them, otherwise they will be held down until the machine reboots. Use lowercase characters as well inside modifiers. For example: to simulate ctrl+c use `<leftCtrlOn>c<leftCtrlOff>`.    

In addition to the special keys, each command to type is treated as a
[configuration template](/docs/templates/configuration-templates.html).
The available variables are:

* `HTTPIP` and `HTTPPort` - The IP and port, respectively of an HTTP server
  that is started serving the directory specified by the `http_directory`
  configuration parameter. If `http_directory` isn't specified, these will
  be blank!

Example boot command. This is actually a working boot command used to start
an Ubuntu 12.04 installer:

```text
[
  "<esc><esc><enter><wait>",
  "/install/vmlinuz noapic ",
  "preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg ",
  "debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
  "hostname={{ .Name }} ",
  "fb=false debconf/frontend=noninteractive ",
  "keyboard-configuration/modelcode=SKIP keyboard-configuration/layout=USA ",
  "keyboard-configuration/variant=USA console-setup/ask_detect=false ",
  "initrd=/install/initrd.gz -- <enter>"
]
```

## Integration Services

Packer will automatically attach the integration services iso as a dvd drive
for the version of HyperV that is running.

## Generation 1 vs Generation 2

Floppy drives are no longer supported by generation 2 machines. This requires you to 
take another approach when dealing with preseed or answer files. Two possible options
are using virtual dvd drives or using the built in web server.

When dealing with Windows you need to enable UEFI drives for generation 2 virtual machines. 

## Creating iso from directory

Programs like mkisofs can be used to create an iso from a directory. 
There is a [windows version of mkisofs](http://opensourcepack.blogspot.co.uk/p/cdrtools.html).

Example powershell script. This is an actually working powershell script used to create a Windows answer iso:

```text
$isoFolder = "answer-iso"
if (test-path $isoFolder){
  remove-item $isoFolder -Force -Recurse
}

if (test-path windows\windows-2012R2-serverdatacenter-amd64\answer.iso){
  remove-item windows\windows-2012R2-serverdatacenter-amd64\answer.iso -Force
}

mkdir $isoFolder

copy windows\windows-2012R2-serverdatacenter-amd64\Autounattend.xml $isoFolder\
copy windows\windows-2012R2-serverdatacenter-amd64\sysprep-unattend.xml $isoFolder\
copy windows\common\set-power-config.ps1 $isoFolder\
copy windows\common\microsoft-updates.ps1 $isoFolder\
copy windows\common\win-updates.ps1 $isoFolder\
copy windows\common\run-sysprep.ps1 $isoFolder\
copy windows\common\run-sysprep.cmd $isoFolder\

$textFile = "$isoFolder\Autounattend.xml" 

$c = Get-Content -Encoding UTF8 $textFile

# Enable UEFI and disable Non EUFI
$c | % { $_ -replace '<!-- Start Non UEFI -->','<!-- Start Non UEFI' } | % { $_ -replace '<!-- Finish Non UEFI -->','Finish Non UEFI -->' } | % { $_ -replace '<!-- Start UEFI compatible','<!-- Start UEFI compatible -->' } | % { $_ -replace 'Finish UEFI compatible -->','<!-- Finish UEFI compatible -->' } | sc -Path $textFile

& .\mkisofs.exe -r -iso-level 4 -UDF -o windows\windows-2012R2-serverdatacenter-amd64\answer.iso $isoFolder

if (test-path $isoFolder){
  remove-item $isoFolder -Force -Recurse
}
```


## Example For Windows Server 2012 R2 Generation 2

Packer config:

```javascript
{
  "builders": [
  {
    "vm_name":"windows2012r2",
    "type": "hyperv-iso",
    "disk_size": 61440,
    "floppy_files": [],
    "secondary_iso_images": [
      "./windows/windows-2012R2-serverdatacenter-amd64/answer.iso"
    ],
    "http_directory": "./windows/common/http/",
    "boot_wait": "0s",
    "boot_command": [
      "a<wait>a<wait>a"
    ],
    "headless": false,
    "iso_url": "http://download.microsoft.com/download/6/2/A/62A76ABB-9990-4EFC-A4FE-C7D698DAEB96/9600.16384.WINBLUE_RTM.130821-1623_X64FRE_SERVER_EVAL_EN-US-IRM_SSS_X64FREE_EN-US_DV5.ISO",
    "iso_checksum_type": "md5",
    "iso_checksum": "458ff91f8abc21b75cb544744bf92e6a",
    "communicator":"winrm",
    "winrm_username": "vagrant",
    "winrm_password": "vagrant",
    "winrm_timeout" : "4h",
    "shutdown_command": "f:\\run-sysprep.cmd",  
    "ram_size": 4096,
    "cpu": 4,
    "generation": 2,
    "switch_name":"LAN",
    "enable_secure_boot":true
  }],
  "provisioners": [{
    "type": "powershell",
    "elevated_user":"vagrant",
    "elevated_password":"vagrant",
    "scripts": [
      "./windows/common/install-7zip.ps1",
      "./windows/common/install-chef.ps1",
      "./windows/common/compile-dotnet-assemblies.ps1",
      "./windows/common/cleanup.ps1",
      "./windows/common/ultradefrag.ps1",
      "./windows/common/sdelete.ps1"
    ]
  }],
  "post-processors": [
    {
      "type": "vagrant",
      "keep_input_artifact": false,
      "output": "{{.Provider}}_windows-2012r2_chef.box"
    }
  ]
}
```

autounattend.xml:

```xml
<?xml version="1.0" encoding="utf-8"?>
<unattend xmlns="urn:schemas-microsoft-com:unattend">
    <settings pass="windowsPE">
        <component name="Microsoft-Windows-International-Core-WinPE" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <SetupUILanguage>
                <UILanguage>en-US</UILanguage>
            </SetupUILanguage>
            <InputLocale>en-US</InputLocale>
            <SystemLocale>en-US</SystemLocale>
            <UILanguage>en-US</UILanguage>
            <UILanguageFallback>en-US</UILanguageFallback>
            <UserLocale>en-US</UserLocale>
        </component>
        <component name="Microsoft-Windows-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <!-- Start Non UEFI -->
            <DiskConfiguration>
                <Disk wcm:action="add">
                    <CreatePartitions>
                        <CreatePartition wcm:action="add">
                            <Type>Primary</Type>
                            <Order>1</Order>
                            <Size>350</Size>
                        </CreatePartition>
                        <CreatePartition wcm:action="add">
                            <Order>2</Order>
                            <Type>Primary</Type>
                            <Extend>true</Extend>
                        </CreatePartition>
                    </CreatePartitions>
                    <ModifyPartitions>
                        <ModifyPartition wcm:action="add">
                            <Active>true</Active>
                            <Format>NTFS</Format>
                            <Label>boot</Label>
                            <Order>1</Order>
                            <PartitionID>1</PartitionID>
                        </ModifyPartition>
                        <ModifyPartition wcm:action="add">
                            <Format>NTFS</Format>
                            <Label>Windows 2012 R2</Label>
                            <Letter>C</Letter>
                            <Order>2</Order>
                            <PartitionID>2</PartitionID>
                        </ModifyPartition>
                    </ModifyPartitions>
                    <DiskID>0</DiskID>
                    <WillWipeDisk>true</WillWipeDisk>
                </Disk>
            </DiskConfiguration>
            <ImageInstall>
                <OSImage>
                    <InstallFrom>
                        <MetaData wcm:action="add">
                            <Key>/IMAGE/NAME </Key>
                            <Value>Windows Server 2012 R2 SERVERSTANDARD</Value>
                        </MetaData>
                    </InstallFrom>
                    <InstallTo>
                        <DiskID>0</DiskID>
                        <PartitionID>2</PartitionID>
                    </InstallTo>
                </OSImage>
            </ImageInstall>
            <!-- Finish Non UEFI -->
            <!-- Start UEFI compatible
            <DiskConfiguration>
                <Disk wcm:action="add">
                    <CreatePartitions>
                        <CreatePartition wcm:action="add">
                            <Order>1</Order>
                            <Size>300</Size>
                            <Type>Primary</Type>
                        </CreatePartition>
                        <CreatePartition wcm:action="add">
                            <Order>2</Order>
                            <Size>100</Size>
                            <Type>EFI</Type>
                        </CreatePartition>
                        <CreatePartition wcm:action="add">
                            <Order>3</Order>
                            <Size>128</Size>
                            <Type>MSR</Type>
                        </CreatePartition>         
                        <CreatePartition wcm:action="add">
                            <Order>4</Order>
                            <Extend>true</Extend> 
                            <Type>Primary</Type>
                        </CreatePartition>
                    </CreatePartitions>
                    <ModifyPartitions>
                        <ModifyPartition wcm:action="add">
                            <Order>1</Order>
                            <PartitionID>1</PartitionID>
                            <Label>WINRE</Label>
                            <Format>NTFS</Format>
                            <TypeID>de94bba4-06d1-4d40-a16a-bfd50179d6ac</TypeID>
                        </ModifyPartition>
                        <ModifyPartition wcm:action="add">
                            <Order>2</Order>
                            <PartitionID>2</PartitionID>
                            <Label>System</Label>
                            <Format>FAT32</Format>
                        </ModifyPartition>
                        <ModifyPartition wcm:action="add">
                            <Order>3</Order>
                            <PartitionID>3</PartitionID>
                        </ModifyPartition>
                        <ModifyPartition wcm:action="add">
                            <Order>4</Order>
                            <PartitionID>4</PartitionID>
                            <Label>Windows</Label>
                            <Format>NTFS</Format>
                        </ModifyPartition>
                    </ModifyPartitions>
                    <DiskID>0</DiskID>
                    <WillWipeDisk>true</WillWipeDisk>
                </Disk>
                <WillShowUI>OnError</WillShowUI>
            </DiskConfiguration>
            <ImageInstall>
                <OSImage>
                    <InstallFrom>
                        <MetaData wcm:action="add">
                            <Key>/IMAGE/NAME </Key>
                            <Value>Windows Server 2012 R2 SERVERSTANDARD</Value>
                        </MetaData>
                    </InstallFrom>
                    <InstallTo>
                        <DiskID>0</DiskID>
                        <PartitionID>4</PartitionID>
                    </InstallTo>
                </OSImage>
            </ImageInstall>
            Finish UEFI compatible -->
            <UserData>
                <!-- Product Key from http://technet.microsoft.com/en-us/library/jj612867.aspx -->
                <ProductKey>
                    <!-- Do not uncomment the Key element if you are using trial ISOs -->
                    <!-- You must uncomment the Key element (and optionally insert your own key) if you are using retail or volume license ISOs -->
                    <!--<Key>D2N9P-3P6X9-2R39C-7RTCD-MDVJX</Key>-->
                    <WillShowUI>OnError</WillShowUI>
                </ProductKey>
                <AcceptEula>true</AcceptEula>
                <FullName>Vagrant</FullName>
                <Organization>Vagrant</Organization>
            </UserData>
        </component>
    </settings>
    <settings pass="specialize">
        <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <OEMInformation>
                <HelpCustomized>false</HelpCustomized>
            </OEMInformation>
            <ComputerName>vagrant-2012r2</ComputerName>
            <TimeZone>Coordinated Universal Time</TimeZone>
            <RegisteredOwner />
        </component>
        <component name="Microsoft-Windows-ServerManager-SvrMgrNc" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <DoNotOpenServerManagerAtLogon>true</DoNotOpenServerManagerAtLogon>
        </component>
        <component name="Microsoft-Windows-IE-ESC" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <IEHardenAdmin>false</IEHardenAdmin>
            <IEHardenUser>false</IEHardenUser>
        </component>
        <component name="Microsoft-Windows-OutOfBoxExperience" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <DoNotOpenInitialConfigurationTasksAtLogon>true</DoNotOpenInitialConfigurationTasksAtLogon>
        </component>
        <component name="Microsoft-Windows-Security-SPP-UX" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <SkipAutoActivation>true</SkipAutoActivation>
        </component>
    </settings>
    <settings pass="oobeSystem">
<!-- Start Setup cache proxy during installation
        <component name="Microsoft-Windows-IE-ClientNetworkProtocolImplementation" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <POLICYProxySettingsPerUser>0</POLICYProxySettingsPerUser>
            <HKLMProxyEnable>true</HKLMProxyEnable>
            <HKLMProxyServer>cache-proxy:3142</HKLMProxyServer>
        </component>  
Finish Setup cache proxy during installation --> 
        <component name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <AutoLogon>
                <Password>
                    <Value>vagrant</Value>
                    <PlainText>true</PlainText>
                </Password>
                <Enabled>true</Enabled>
                <Username>vagrant</Username>
            </AutoLogon>
            <FirstLogonCommands>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c powershell -Command "Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Force"</CommandLine>
                    <Description>Set Execution Policy 64 Bit</Description>
                    <Order>1</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>C:\Windows\SysWOW64\cmd.exe /c powershell -Command "Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Force"</CommandLine>
                    <Description>Set Execution Policy 32 Bit</Description>
                    <Order>2</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c winrm quickconfig -q</CommandLine>
                    <Description>winrm quickconfig -q</Description>
                    <Order>3</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c winrm quickconfig -transport:http</CommandLine>
                    <Description>winrm quickconfig -transport:http</Description>
                    <Order>4</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c winrm set winrm/config @{MaxTimeoutms="1800000"}</CommandLine>
                    <Description>Win RM MaxTimoutms</Description>
                    <Order>5</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c winrm set winrm/config/winrs @{MaxMemoryPerShellMB="300"}</CommandLine>
                    <Description>Win RM MaxMemoryPerShellMB</Description>
                    <Order>6</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c winrm set winrm/config/service @{AllowUnencrypted="true"}</CommandLine>
                    <Description>Win RM AllowUnencrypted</Description>
                    <Order>7</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c winrm set winrm/config/service/auth @{Basic="true"}</CommandLine>
                    <Description>Win RM auth Basic</Description>
                    <Order>8</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c winrm set winrm/config/client/auth @{Basic="true"}</CommandLine>
                    <Description>Win RM client auth Basic</Description>
                    <Order>9</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c winrm set winrm/config/listener?Address=*+Transport=HTTP @{Port="5985"} </CommandLine>
                    <Description>Win RM listener Address/Port</Description>
                    <Order>10</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c netsh advfirewall firewall set rule group="remote administration" new enable=yes </CommandLine>
                    <Description>Win RM adv firewall enable</Description>
                    <Order>11</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c netsh advfirewall firewall add rule name="WinRM 5985" protocol=TCP dir=in localport=5985 action=allow</CommandLine>
                    <Description>Win RM port open</Description>
                    <Order>12</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c netsh advfirewall firewall add rule name="WinRM 5986" protocol=TCP dir=in localport=5986 action=allow</CommandLine>
                    <Description>Win RM port open</Description>
                    <Order>13</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c net stop winrm </CommandLine>
                    <Description>Stop Win RM Service </Description>
                    <Order>14</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c sc config winrm start= disabled</CommandLine>
                    <Description>Win RM Autostart</Description>
                    <Order>15</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>%SystemRoot%\System32\reg.exe ADD HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\Advanced\ /v HideFileExt /t REG_DWORD /d 0 /f</CommandLine>
                    <Order>16</Order>
                    <Description>Show file extensions in Explorer</Description>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>%SystemRoot%\System32\reg.exe ADD HKCU\Console /v QuickEdit /t REG_DWORD /d 1 /f</CommandLine>
                    <Order>17</Order>
                    <Description>Enable QuickEdit mode</Description>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>%SystemRoot%\System32\reg.exe ADD HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\Advanced\ /v Start_ShowRun /t REG_DWORD /d 1 /f</CommandLine>
                    <Order>18</Order>
                    <Description>Show Run command in Start Menu</Description>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>%SystemRoot%\System32\reg.exe ADD HKCU\SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\Advanced\ /v StartMenuAdminTools /t REG_DWORD /d 1 /f</CommandLine>
                    <Order>19</Order>
                    <Description>Show Administrative Tools in Start Menu</Description>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>%SystemRoot%\System32\reg.exe ADD HKLM\SYSTEM\CurrentControlSet\Control\Power\ /v HibernateFileSizePercent /t REG_DWORD /d 0 /f</CommandLine>
                    <Order>20</Order>
                    <Description>Zero Hibernation File</Description>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>%SystemRoot%\System32\reg.exe ADD HKLM\SYSTEM\CurrentControlSet\Control\Power\ /v HibernateEnabled /t REG_DWORD /d 0 /f</CommandLine>
                    <Order>21</Order>
                    <Description>Disable Hibernation Mode</Description>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c wmic useraccount where "name='vagrant'" set PasswordExpires=FALSE</CommandLine>
                    <Order>22</Order>
                    <Description>Disable password expiration for vagrant user</Description>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c winrm set winrm/config/winrs @{MaxShellsPerUser="30"}</CommandLine>
                    <Description>Win RM MaxShellsPerUser</Description>
                    <Order>23</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c winrm set winrm/config/winrs @{MaxProcessesPerShell="25"}</CommandLine>
                    <Description>Win RM MaxProcessesPerShell</Description>
                    <Order>24</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>%SystemRoot%\System32\reg.exe ADD "HKLM\System\CurrentControlSet\Services\Netlogon\Parameters" /v DisablePasswordChange /t REG_DWORD /d 1 /f</CommandLine>
                    <Description>Turn off computer password</Description>
                    <Order>25</Order>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c netsh advfirewall firewall add rule name="ICMP Allow incoming V4 echo request" protocol=icmpv4:8,any dir=in action=allow</CommandLine>
                    <Description>ICMP open for ping</Description>
                    <Order>26</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <!-- WITH WINDOWS UPDATES -->
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c IF EXIST a:\set-power-config.ps1 (C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -File a:\set-power-config.ps1) ELSE (C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -File f:\set-power-config.ps1)</CommandLine>
                    <Order>97</Order>
                    <Description>Turn off all power saving and timeouts</Description>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c IF EXIST a:\microsoft-updates.ps1 (C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -File a:\microsoft-updates.ps1) ELSE (C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -File f:\microsoft-updates.ps1)</CommandLine>
                    <Order>98</Order>
                    <Description>Enable Microsoft Updates</Description>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <SynchronousCommand wcm:action="add">
                    <CommandLine>cmd.exe /c IF EXIST a:\win-updates.ps1 (C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -File a:\win-updates.ps1) ELSE (C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe -File f:\win-updates.ps1)</CommandLine>
                    <Description>Install Windows Updates</Description>
                    <Order>100</Order>
                    <RequiresUserInput>true</RequiresUserInput>
                </SynchronousCommand>
                <!-- END WITH WINDOWS UPDATES -->
            </FirstLogonCommands>
            <OOBE>
                <HideEULAPage>true</HideEULAPage>
                <HideLocalAccountScreen>true</HideLocalAccountScreen>
                <HideOEMRegistrationScreen>true</HideOEMRegistrationScreen>
                <HideOnlineAccountScreens>true</HideOnlineAccountScreens>
                <HideWirelessSetupInOOBE>true</HideWirelessSetupInOOBE>
                <NetworkLocation>Work</NetworkLocation>
                <ProtectYourPC>1</ProtectYourPC>
            </OOBE>
            <UserAccounts>
                <AdministratorPassword>
                    <Value>vagrant</Value>
                    <PlainText>true</PlainText>
                </AdministratorPassword>
                <LocalAccounts>
                    <LocalAccount wcm:action="add">
                        <Password>
                            <Value>vagrant</Value>
                            <PlainText>true</PlainText>
                        </Password>
                        <Group>administrators</Group>
                        <DisplayName>Vagrant</DisplayName>
                        <Name>vagrant</Name>
                        <Description>Vagrant User</Description>
                    </LocalAccount>
                </LocalAccounts>
            </UserAccounts>
            <RegisteredOwner />
            <TimeZone>Coordinated Universal Time</TimeZone>
        </component>
    </settings>
    <settings pass="offlineServicing">
        <component name="Microsoft-Windows-LUA-Settings" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <EnableLUA>false</EnableLUA>
        </component>
    </settings>
    <cpi:offlineImage cpi:source="wim:c:/projects/baseboxes/9600.16384.winblue_rtm.130821-1623_x64fre_server_eval_en-us-irm_sss_x64free_en-us_dv5_slipstream/sources/install.wim#Windows Server 2012 R2 SERVERDATACENTER" xmlns:cpi="urn:schemas-microsoft-com:cpi" />
</unattend>

```

sysprep-unattend.xml:

```text
<?xml version="1.0" encoding="utf-8"?>
<unattend xmlns="urn:schemas-microsoft-com:unattend">
    <settings pass="generalize">
        <component language="neutral" name="Microsoft-Windows-Security-SPP" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <SkipRearm>1</SkipRearm>
        </component>
    </settings>
    <settings pass="oobeSystem">
<!-- Setup proxy after sysprep 
       <component name="Microsoft-Windows-IE-ClientNetworkProtocolImplementation" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" language="neutral" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <POLICYProxySettingsPerUser>1</POLICYProxySettingsPerUser>
            <HKLMProxyEnable>false</HKLMProxyEnable>
            <HKLMProxyServer>cache-proxy:3142</HKLMProxyServer>
        </component>
Finish proxy after sysprep -->  
        <component language="neutral" name="Microsoft-Windows-International-Core" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <InputLocale>0809:00000809</InputLocale>
            <SystemLocale>en-GB</SystemLocale>
            <UILanguage>en-US</UILanguage>
            <UILanguageFallback>en-US</UILanguageFallback>
            <UserLocale>en-GB</UserLocale>
        </component>
        <component language="neutral" name="Microsoft-Windows-Shell-Setup" processorArchitecture="amd64" publicKeyToken="31bf3856ad364e35" versionScope="nonSxS" xmlns:wcm="http://schemas.microsoft.com/WMIConfig/2002/State" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
            <OOBE>
                <HideEULAPage>true</HideEULAPage>
                <HideOEMRegistrationScreen>true</HideOEMRegistrationScreen>
                <HideOnlineAccountScreens>true</HideOnlineAccountScreens>
                <HideWirelessSetupInOOBE>true</HideWirelessSetupInOOBE>
                <NetworkLocation>Work</NetworkLocation>
                <ProtectYourPC>1</ProtectYourPC>
                <SkipUserOOBE>true</SkipUserOOBE>
                <SkipMachineOOBE>true</SkipMachineOOBE>
            </OOBE>
            <UserAccounts>
                <AdministratorPassword>
                    <Value>vagrant</Value>
                    <PlainText>true</PlainText>
                </AdministratorPassword>
                <LocalAccounts>
                    <LocalAccount wcm:action="add">
                        <Password>
                            <Value>vagrant</Value>
                            <PlainText>true</PlainText>
                        </Password>
                        <Group>administrators</Group>
                        <DisplayName>Vagrant</DisplayName>
                        <Name>vagrant</Name>
                        <Description>Vagrant User</Description>
                    </LocalAccount>
                </LocalAccounts>
            </UserAccounts>
            <DisableAutoDaylightTimeSet>true</DisableAutoDaylightTimeSet>
            <TimeZone>Coordinated Universal Time</TimeZone>
            <VisualEffects>
                <SystemDefaultBackgroundColor>2</SystemDefaultBackgroundColor>
            </VisualEffects>
        </component>
    </settings>
</unattend>
```

## Example For Ubuntu Vivid Generation 2

Packer config:

```javascript
{
  "builders": [
  {
    "vm_name":"ubuntu-vivid",
    "type": "hyperv-iso",
    "disk_size": 61440,
    "headless": false,
    "iso_url": "http://releases.ubuntu.com/15.04/ubuntu-15.04-server-amd64.iso",
    "iso_checksum_type": "sha1",
    "iso_checksum": "D10248965C2C749DF6BCCE9F2F90F16A2E75E843",
    "communicator":"ssh",
    "ssh_username": "vagrant",
    "ssh_password": "vagrant",
    "ssh_timeout" : "4h",
    "http_directory": "./linux/ubuntu/http/",
    "boot_wait": "5s",
    "boot_command": [
      "<esc><esc><enter><wait>",
      "/install/vmlinuz ",
      "preseed/url=http://{{.HTTPIP}}:{{.HTTPPort}}/preseed.cfg ",
      "debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
      "hostname={{.Name}} ",
      "fb=false debconf/frontend=noninteractive ",
      "keyboard-configuration/modelcode=SKIP keyboard-configuration/layout=USA ",
      "keyboard-configuration/variant=USA console-setup/ask_detect=false ",
      "initrd=/install/initrd.gz -- <enter>"
    ],
    "shutdown_command": "echo 'vagrant' | sudo -S -E shutdown -P now",
    "ram_size": 4096,
    "cpu": 4,
    "generation": 1,
    "switch_name":"LAN"
  }],
  "provisioners": [{
    "type": "shell",
    "execute_command": "echo 'vagrant' | sudo -S -E sh {{.Path}}",
    "scripts": [
      "./linux/ubuntu/update.sh",
      "./linux/ubuntu/network.sh",
      "./linux/common/vagrant.sh",
      "./linux/common/chef.sh",
      "./linux/common/motd.sh",
      "./linux/ubuntu/cleanup.sh"
    ]
  }],
  "post-processors": [
    {
      "type": "vagrant",
      "keep_input_artifact": true,
      "output": "{{.Provider}}_ubuntu-15.04_chef.box"
    }
  ]
}
```

preseed.cfg:

```text
## Options to set on the command line
d-i debian-installer/locale string en_US.utf8
d-i console-setup/ask_detect boolean false
d-i console-setup/layout string us

d-i netcfg/get_hostname string unassigned-hostname
d-i netcfg/get_domain string unassigned-domain

d-i time/zone string UTC
d-i clock-setup/utc-auto boolean true
d-i clock-setup/utc boolean true

d-i kbd-chooser/method select American English

d-i netcfg/wireless_wep string

d-i base-installer/kernel/override-image string linux-server

d-i debconf debconf/frontend select Noninteractive

d-i pkgsel/install-language-support boolean false
tasksel tasksel/first multiselect standard, ubuntu-server

d-i partman-auto/method string lvm

d-i partman-lvm/confirm boolean true
d-i partman-lvm/device_remove_lvm boolean true
d-i partman-auto/choose_recipe select atomic

d-i partman/confirm_write_new_label boolean true
d-i partman/confirm_nooverwrite boolean true
d-i partman/choose_partition select finish
d-i partman/confirm boolean true

# Write the changes to disks and configure LVM?
d-i partman-lvm/confirm boolean true
d-i partman-lvm/confirm_nooverwrite boolean true
d-i partman-auto-lvm/guided_size string max

# Default user
d-i passwd/user-fullname string vagrant
d-i passwd/username string vagrant
d-i passwd/user-password password vagrant
d-i passwd/user-password-again password vagrant
d-i user-setup/encrypt-home boolean false
d-i user-setup/allow-password-weak boolean true

# Minimum packages (see postinstall.sh)
d-i pkgsel/include string openssh-server ntp

# Upgrade packages after debootstrap? (none, safe-upgrade, full-upgrade)
# (note: set to none for speed)
d-i pkgsel/upgrade select none

d-i grub-installer/only_debian boolean true
d-i grub-installer/with_other_os boolean true
d-i finish-install/reboot_in_progress note

d-i pkgsel/update-policy select none

choose-mirror-bin mirror/http/proxy string

#d-i mirror/http/proxy string http://apt-cacher:3142/
```