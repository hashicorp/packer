---
description: |
    This VSphere Packer builder is able to create VSphere virtual machines from an ISO
    file as a source. It currently supports building virtual machines on hosts
    running VSphere Fusion for OS X, VSphere Workstation for Linux and Windows, and
    VSphere Player on Linux. It can also build machines directly on VSphere vSphere
    Hypervisor using SSH as opposed to the vSphere API.
layout: docs
page_title: VSphere Builder from ISO
...

# VSphere Builder (from ISO)

Type: `VSphere-iso`

This VSphere Packer builder is able to create VSphere virtual machines from an ISO
file as a source. It currently supports building virtual machines directly on [VSphere vSphere
Hypervisor](https://www.VSphere.com/products/vsphere-hypervisor/) using the vSphere API.

The builder builds a virtual machine by creating a new virtual machine from
scratch, booting it, installing an OS, provisioning software within the OS, then
shutting it down and export it back to packer machine. The result of the VSphere builder is a directory containing all
the files necessary to import to a Vsphere product.

Packer will output VNC connection information in case you
    need to connect to the console to debug the build process.

## Basic Example

Here is a basic example. This example is not functional. It will start the OS
installer but then fail because we don't provide the preseed file for Ubuntu to
self-install. Still, the example serves to show the basic configuration:

``` {.javascript}
{
  "type": "vsphere-iso",
  "iso_url": "http://old-releases.ubuntu.com/releases/precise/ubuntu-12.04.2-server-amd64.iso",
  "iso_checksum": "af5f788aee1b32c4b2634734309cc9e9",
  "iso_checksum_type": "md5",
  "ssh_username": "packer",
  "shutdown_command": "shutdown -P now"
}
```

## Configuration Reference

There are many configuration options available for the VSphere builder. They are
organized below into two categories: required and optional. Within each
category, the available options are alphabetized and described.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `iso_checksum` (string) - The checksum for the OS ISO file. Because ISO
    files are so large, this is required and Packer will verify it prior to
    booting a virtual machine with the ISO attached. The type of the checksum is
    specified with `iso_checksum_type`, documented below. At least one of
    `iso_checksum` and `iso_checksum_url` must be defined. This has precedence
    over `iso_checksum_url` type.

-   `iso_checksum_type` (string) - The type of the checksum specified in
    `iso_checksum`. Valid values are "none", "md5", "sha1", "sha256", or
    "sha512" currently. While "none" will skip checksumming, this is not
    recommended since ISO files are generally large and corruption does happen
    from time to time.

-   `iso_checksum_url` (string) - A URL to a GNU or BSD style checksum file
    containing a checksum for the OS ISO file. At least one of `iso_checksum`
    and `iso_checksum_url` must be defined. This will be ignored if
    `iso_checksum` is non empty.

-   `iso_url` (string) - A URL to the ISO containing the installation image.
    This URL can be either an HTTP URL or a file URL (or path to a file). If
    this is an HTTP URL, Packer will download it and cache it between runs.

-   `ssh_username` (string) - The username to use to SSH into the machine once
    the OS is installed.

### Optional:

-   `annotation` (string) - The value of this will be placed in the annotation field of the virtual machine.

-   `boot_command` (array of strings) - This is an array of commands to type
    when the virtual machine is first booted. The goal of these commands should
    be to type just enough to initialize the operating system installer. Special
    keys can be typed as well, and are covered in the section below on the
    boot command. If this is not specified, it is assumed the installer will
    start itself.

-   `boot_wait` (string) - The time to wait after booting the initial virtual
    machine before typing the `boot_command`. The value of this should be
    a duration. Examples are "5s" and "1m30s" which will cause Packer to wait
    five seconds and one minute 30 seconds, respectively. If this isn't
    specified, the default is 10 seconds.

-   `cpu` (integer) - The number of CPU. Default to "1".

-   `disk_additional_size` (array of integers) - The size(s) of any additional
    hard disks for the VM in megabytes. If this is not specified then the VM
    will only contain a primary hard disk. The builder uses expandable, not
    fixed-size virtual hard disks, so the actual file representing the disk will
    not use the full size unless it is full.

-   `disk_size` (integer) - The size of the hard disk for the VM in megabytes.
    The builder uses expandable, not fixed-size virtual hard disks, so the
    actual file representing the disk will not use the full size unless it
    is full. By default this is set to 40,000 (about 40 GB).

-   `disk_thick` (bool) - The type of VSphere virtual disks to create. The
    default is "false", which corresponds to a growable virtual disk (thin provisioning).If "true" the disks will be fixed size (thick provisioning).

-   `floppy_files` (array of strings) - A list of files to place onto a floppy
    disk that is attached when the VM is booted. This is most useful for
    unattended Windows installs, which look for an `Autounattend.xml` file on
    removable media. By default, no floppy will be attached. All files listed in
    this setting get placed into the root directory of the floppy and the floppy
    is attached as the first floppy device. Currently, no support exists for
    creating sub-directories on the floppy. Wildcard characters (\*, ?,
    and \[\]) are allowed. Directory names are also allowed, which will add all
    the files found in the directory to the floppy.

-   `floppy_dirs` (array of strings) - A list of directories to place onto
    the floppy disk recursively. This is similar to the `floppy_files` option
    except that the directory structure is preserved. This is useful for when
    your floppy disk includes drivers or if you just want to organize it's
    contents as a hierarchy. Wildcard characters (\*, ?, and \[\]) are allowed.

-   `format` (string) - Either "ovf", "ova" or "vmx", this specifies the output
    format of the exported virtual machine. This defaults to "ovf".
    Before using this option, you need to install `ovftool`.

-   `guest_os_type` (string) - The guest OS type being installed. This will be
    set in the virtual machine VMX file. By default this is "otherGuest". By specifying a more
    specific OS type, VSphere may perform some optimizations or virtual hardware
    changes to better support the operating system running in the
    virtual machine. **Note:** This value are different from vmware-iso builder. Search for `vim.vm.GuestOsDescriptor.GuestOsIdentifier` in VMware API documentation for a comprehensive list.

-   `http_directory` (string) - Path to a directory to serve using an
    HTTP server. The files in this directory will be available over HTTP that
    will be requestable from the virtual machine. This is useful for hosting
    kickstart files and so on. By default this is "", which means no HTTP server
    will be started. The address and port of the HTTP server will be available
    as variables in `boot_command`. This is covered in more detail below.

-   `http_port_min` and `http_port_max` (integer) - These are the minimum and
    maximum port to use for the HTTP server started to serve the
    `http_directory`. Because Packer often runs in parallel, Packer will choose
    a randomly available port in this range to run the HTTP server. If you want
    to force the HTTP server to be on one port, make this minimum and maximum
    port the same. By default the values are 8000 and 9000, respectively.

-   `insecure` (bool) - If `true` access to VSphere API over HTTPS will not verify the server certificate validity. The value is `false` by default

-   `iso_target_extension` (string) - The extension of the iso file after
    download. This defaults to "iso".

-   `iso_target_path` (string) - The path where the iso should be saved after
    download. By default will go in the packer cache, with a hash of the
    original filename as its name.

-   `iso_urls` (array of strings) - Multiple URLs for the ISO to download.
    Packer will try these in order. If anything goes wrong attempting to
    download or while downloading a single URL, it will move on to the next. All
    URLs must point to the same file (same checksum). By default this is empty
    and `iso_url` is used. Only one of `iso_url` or `iso_urls` can be specified.

-   `keep_registered` (boolean) - Set this to `true` if you would like to keep
    the VM registered with the remote Vsphere server. This is convenient if you
    use packer to provision VMs on Vsphere and don't want to use ovftool to
    deploy the resulting artifact (VMX or OVA or whatever you used as `format`).
    Defaults to `false`.

-   `mem_size` (integer) - The size(s) of memory in MB. Defaults to "1024".

-   `network_adapter` (string) - Type of the first ethernet card. Defaults to "e1000".

-   `network_name` (string) - Name of network for the first ethernet card. Defaults to "VM Network".

-   `output_directory` (string) - This is the path to the directory where the
    resulting virtual machine will be created. This may be relative or absolute.
    If relative, the path is relative to the working directory when `packer`
    is executed. This directory must not exist or be empty prior to running
    the builder. By default this is "output-BUILDNAME" where "BUILDNAME" is the
    name of the build.

-   `ovftool_options` (array of string) - Specifies the additionnal options to
    be used at export phases by `ovftool`.

-   `remote_cache_datastore` (string) - The path to the datastore where
    supporting files will be stored during the build on the remote machine. By
    default this is the same as the `remote_datastore` option.

-   `remote_cache_directory` (string) - The path where the ISO and/or floppy
    files will be stored during the build on the remote machine. The path is
    relative to the `remote_cache_datastore` on the remote machine. By default
    this is "packer\_cache".

-   `remote_cluster` (string) - The name of cluster (if relevant) in which the host is.

-   `remote_datacenter` (string) - The name of datacenter in which the host is. On ESXi not managed by Vcenter the default value of `ha-datacenter` is sufficient.

-   `remote_datastore` (string) - The path to the datastore where the resulting
    VM will be stored when it is built on the remote machine. By default this
    is "datastore1".

-   `remote_folder` (string) - The logical path of the
    VM on the remote machine.

-   `remote_host` (string) - The host of the remote machine used for access.

-   `remote_password` (string) - The SSH password for the user used to access
    the remote machine. By default this is empty.

-   `remote_resource_pool` (string) - The resource pool of the
    VM on the remote machine.

-   `remote_username` (string) - The username for the SSH user that will access
    the remote machine. This is required if `remote_type` is enabled.

-   `shutdown_command` (string) - The command to use to gracefully shut down the
    machine once all the provisioning is done. By default this is an empty
    string, which tells Packer to just forcefully shut down the machine.

-   `shutdown_timeout` (string) - The amount of time to wait after executing the
    `shutdown_command` for the virtual machine to actually shut down. If it
    doesn't shut down in this time, it is an error. By default, the timeout is
    "5m", or five minutes.


-   `skip_export` (boolean) - Defaults to `false`. When enabled, Packer will
    not export the VM. Useful if the build output is not the resultant image,
    but created inside the VM.

-   `vcenter` (string) - Name of vcenter managing the host used to generate the VM. If not provided, it will default to host name.

-   `vm_name` (string) - This is the name of the VMX file for the new virtual
    machine, without the file extension. By default this is "packer-BUILDNAME",
    where "BUILDNAME" is the name of the build.

-   `vmx_data` (object of key/value strings) - Arbitrary key/values to enter
    into the virtual machine VMX file. This is for advanced users who want to
    set properties other than memory and CPU such as security policy etc. **Note:** The values inserted with this option are not removed when `vmx_data_post` is processed (this behavior is different than vmware-iso).

-   `vmx_data_post` (object of key/value strings) - Identical to `vmx_data`,
    except that it is run after the virtual machine is shutdown, and before the
    virtual machine is exported.

-   `vnc_bind_address` (string / IP address) - The IP address that should be binded
    to for VNC. By default packer will use 127.0.0.1 for this. If you wish to bind
    to all interfaces use 0.0.0.0

-   `vnc_disable_password` (boolean) - Don't auto-generate a VNC password that is
    used to secure the VNC communication with the VM.

-   `vnc_port_min` and `vnc_port_max` (integer) - The minimum and maximum port
    to use for VNC access to the virtual machine. The builder uses VNC to type
    the initial `boot_command`. Because Packer generally runs in parallel,
    Packer uses a randomly chosen port in this range that appears available. By
    default this is 5900 to 6000. The minimum and maximum ports are inclusive.

## Boot Command

The `boot_command` configuration is very important: it specifies the keys to
type when the virtual machine is first booted in order to start the OS
installer. This command is typed after `boot_wait`, which gives the virtual
machine some time to actually load the ISO.

As documented above, the `boot_command` is an array of strings. The strings are
all typed in sequence. It is an array only to improve readability within the
template.

The boot command is "typed" character for character over a VNC connection to the
machine, simulating a human actually typing the keyboard.

-> Keystrokes are typed as separate key up/down events over VNC with a
   default 100ms delay. The delay alleviates issues with latency and CPU
   contention. For local builds you can tune this delay by specifying
   e.g. `PACKER_KEY_INTERVAL=10ms` to speed through the boot command.

There are a set of special keys available. If these are in your boot
command, they will be replaced by the proper key:

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

When using modifier keys `ctrl`, `alt`, `shift` ensure that you release them,
otherwise they will be held down until the machine reboots. Use lowercase
characters as well inside modifiers.

For example: to simulate ctrl+c use `<leftCtrlOn>c<leftCtrlOff>`.

In addition to the special keys, each command to type is treated as a
[configuration template](/docs/templates/configuration-templates.html). The
available variables are:

-   `HTTPIP` and `HTTPPort` - The IP and port, respectively of an HTTP server
    that is started serving the directory specified by the `http_directory`
    configuration parameter. If `http_directory` isn't specified, these will be
    blank!

Example boot command. This is actually a working boot command used to start an
Ubuntu 12.04 installer:

``` {.text}
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


### Using a Floppy for Linux kickstart file or preseed

Depending on your network configuration, it may be difficult to use packer's
built-in HTTP server with VSphere. Instead, you can provide a kickstart or preseed
file by attaching a floppy disk. An example below, based on RHEL:

``` {.javascript}
{
  "builders": [
    {
      "type":"VSphere-iso",
      "floppy_files": [
        "folder/ks.cfg"
      ],
      "boot_command": "<tab> text ks=floppy <enter><wait>"
    }
  ]
}
```

It's also worth noting that `ks=floppy` has been deprecated.  Later versions of the Anaconda installer (used in RHEL/CentOS 7 and Fedora) may require a different syntax to source a kickstart file from a mounted floppy image.

``` {.javascript}
{
  "builders": [
    {
      "type":"VSphere-iso",
      "floppy_files": [
        "folder/ks.cfg"
      ],
      "boot_command": "<tab> inst.text inst.ks=hd:fd0:/ks.cfg <enter><wait>"
    }
  ]
}
```
