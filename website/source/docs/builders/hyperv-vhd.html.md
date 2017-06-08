---
description: |-
  The Hyper-V Packer builder is able to create Hyper-V virtual machines and export them.
layout: "docs"
page_title: "Hyper-V Builder (from an VHD)"
---

# Hyper-V Builder (from an VHD)

Type: `hyperv-vhd`

The Hyper-V Packer builder is able to create [Hyper-V](https://www.microsoft.com/en-us/server-cloud/solutions/virtualization.aspx)
virtual machines and export them, based on an existing hyper-v export.

The builder builds a virtual machine by importing an existing export, copying its 
files, booting it, provisioning software within the OS, then shutting it down. 
The result of the Hyper-V builder is a directory containing all the files 
necessary to run the virtual machine portably.

## Basic Example

Here is a basic example:

```javascript
{
  "type": "hyperv-vhd",
  "source_dir": "./output-hyperv-iso",
  "vm_name": "ubuntu-xenial",
  "ssh_username": "packer",
  "ssh_password": "packer",
  "shutdown_command": "echo 'packer' | sudo -S shutdown -P now"
}
```

It is important to add a `shutdown_command`. By default Packer halts the
virtual machine and the file system may not be sync'd. Thus, changes made in a
provisioner might not be saved.

## Configuration Reference

There are many configuration options available for the Hyper-V builder.
They are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html)
can be configured for this builder.

### Required:

-   `source_dir` (string) - The source directory of the existing export, which
    should be an output of hyper-v iso builder or a manually export from hyper-v
    manager. The directory should contains two sub-directories, one for virtual 
    hard disk file and another one for VM configuration file. The configuration
    filename should be a GUID (virtual machine ID) with .xml extension (Windows 
    Server 2012) or .vmcx extension (Windows Server 2016).
    
    The builder will try to find the **first match** configuration file and import
    virtual machine by PowerShell commands.

### Optional:

-   `generation` (integer) - The Hyper-V generation for the virtual machine, which
    **should be consistent with the source hyper-v export**. By default, this is 1. 
    Generation 2 Hyper-V virtual machines do not support floppy drives. In this 
    scenario use `secondary_iso_images` instead. Hard drives and dvd drives will 
    also be scsi and not ide. 

-   `switch_name` (string) - The name of the switch to connect the virtual machine 
    to, which **should be consistent with the source hyper-v export**. Be defaulting
    this to an empty string, Packer will try to determine the switch to use by 
    looking for external switch that is up and running.

-   `boot_command` (array of strings) - This is an array of commands to type
    when the virtual machine is first booted. If this is not specified, it is 
    assumed the virtual machine will start itself.

-   `boot_wait` (string) - The time to wait after booting the initial virtual
    machine before typing the `boot_command`. The value of this should be
    a duration. Examples are "5s" and "1m30s" which will cause Packer to wait
    five seconds and one minute 30 seconds, respectively. If this isn't specified,
    the default is 10 seconds.

-   `floppy_files` (array of strings) - A list of files to place onto a floppy
    disk that is attached when the VM is booted. This is most useful
    for unattended Windows installs, which look for an `Autounattend.xml` file
    on removable media. By default, no floppy will be attached. All files
    listed in this setting get placed into the root directory of the floppy
    and the floppy is attached as the first floppy device. Currently, no
    support exists for creating sub-directories on the floppy. Wildcard
    characters (*, ?, and []) are allowed. Directory names are also allowed,
    which will add all the files found in the directory to the floppy.

-   `guest_additions_mode` (string) - How should guest additions be installed.
    If value `attach` then attach iso image with by specified by `guest_additions_path`.
    Otherwise guest additions is not installed.

-   `guest_additions_path` (string) - The path to the iso image for guest additions.

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

-   `switch_vlan_id` (string) - This is the vlan of the virtual switch's network card. 
    By default none is set. If none is set then a vlan is not set on the switch's network card.
    If this value is set it should match the vlan specified in by `vlan_id`.

-   `vlan_id` (string) - This is the vlan of the virtual machine's network card for the new virtual
    machine. By default none is set. If none is set then vlans are not set on the virtual machine's 
    network card.

-   `vm_name` (string) - This is the name of the virtual machine for the new virtual
    machine, without the file extension. By default this is "packer-BUILDNAME",
    where "BUILDNAME" is the name of the build.

