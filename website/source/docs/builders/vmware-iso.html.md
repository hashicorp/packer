---
description: |
    This VMware Packer builder is able to create VMware virtual machines from an ISO
    file as a source. It currently supports building virtual machines on hosts
    running VMware Fusion for OS X, VMware Workstation for Linux and Windows, and
    VMware Player on Linux. It can also build machines directly on VMware vSphere
    Hypervisor using SSH as opposed to the vSphere API.
layout: docs
page_title: VMware Builder from ISO
...

# VMware Builder (from ISO)

Type: `vmware-iso`

This VMware Packer builder is able to create VMware virtual machines from an ISO
file as a source. It currently supports building virtual machines on hosts
running [VMware Fusion](https://www.vmware.com/products/fusion/overview.html) for
OS X, [VMware
Workstation](https://www.vmware.com/products/workstation/overview.html) for Linux
and Windows, and [VMware Player](https://www.vmware.com/products/player/) on
Linux. It can also build machines directly on [VMware vSphere
Hypervisor](https://www.vmware.com/products/vsphere-hypervisor/) using SSH as
opposed to the vSphere API.

The builder builds a virtual machine by creating a new virtual machine from
scratch, booting it, installing an OS, provisioning software within the OS, then
shutting it down. The result of the VMware builder is a directory containing all
the files necessary to run the virtual machine.

## Basic Example

Here is a basic example. This example is not functional. It will start the OS
installer but then fail because we don't provide the preseed file for Ubuntu to
self-install. Still, the example serves to show the basic configuration:

``` {.javascript}
{
  "type": "vmware-iso",
  "iso_url": "http://old-releases.ubuntu.com/releases/precise/ubuntu-12.04.2-server-amd64.iso",
  "iso_checksum": "af5f788aee1b32c4b2634734309cc9e9",
  "iso_checksum_type": "md5",
  "ssh_username": "packer",
  "shutdown_command": "shutdown -P now"
}
```

## Configuration Reference

There are many configuration options available for the VMware builder. They are
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

-   `disk_additional_size` (array of integers) - The size(s) of any additional
    hard disks for the VM in megabytes. If this is not specified then the VM
    will only contain a primary hard disk. The builder uses expandable, not
    fixed-size virtual hard disks, so the actual file representing the disk will
    not use the full size unless it is full.

-   `disk_size` (integer) - The size of the hard disk for the VM in megabytes.
    The builder uses expandable, not fixed-size virtual hard disks, so the
    actual file representing the disk will not use the full size unless it
    is full. By default this is set to 40,000 (about 40 GB).

-   `disk_type_id` (string) - The type of VMware virtual disk to create. The
    default is "1", which corresponds to a growable virtual disk split in
    2GB files. This option is for advanced usage, modify only if you know what
    you're doing. For more information, please consult the [Virtual Disk Manager
    User's Guide](https://www.vmware.com/pdf/VirtualDiskManager.pdf) for desktop
    VMware clients. For ESXi, refer to the proper ESXi documentation.

-   `floppy_files` (array of strings) - A list of files to place onto a floppy
    disk that is attached when the VM is booted. This is most useful for
    unattended Windows installs, which look for an `Autounattend.xml` file on
    removable media. By default, no floppy will be attached. All files listed in
    this setting get placed into the root directory of the floppy and the floppy
    is attached as the first floppy device. Currently, no support exists for
    creating sub-directories on the floppy. Wildcard characters (\*, ?,
    and \[\]) are allowed. Directory names are also allowed, which will add all
    the files found in the directory to the floppy.

-   `fusion_app_path` (string) - Path to "VMware Fusion.app". By default this is
    "/Applications/VMware Fusion.app" but this setting allows you to
    customize this.

-   `guest_os_type` (string) - The guest OS type being installed. This will be
    set in the VMware VMX. By default this is "other". By specifying a more
    specific OS type, VMware may perform some optimizations or virtual hardware
    changes to better support the operating system running in the
    virtual machine.

-   `headless` (boolean) - Packer defaults to building VMware virtual machines
    by launching a GUI that shows the console of the machine being built. When
    this value is set to true, the machine will start without a console. For
    VMware machines, Packer will output VNC connection information in case you
    need to connect to the console to debug the build process.

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

-   `iso_target_path` (string) - The path where the iso should be saved after
    download. By default will go in the packer cache, with a hash of the
    original filename as its name.

-   `iso_urls` (array of strings) - Multiple URLs for the ISO to download.
    Packer will try these in order. If anything goes wrong attempting to
    download or while downloading a single URL, it will move on to the next. All
    URLs must point to the same file (same checksum). By default this is empty
    and `iso_url` is used. Only one of `iso_url` or `iso_urls` can be specified.

-   `output_directory` (string) - This is the path to the directory where the
    resulting virtual machine will be created. This may be relative or absolute.
    If relative, the path is relative to the working directory when `packer`
    is executed. This directory must not exist or be empty prior to running
    the builder. By default this is "output-BUILDNAME" where "BUILDNAME" is the
    name of the build.

-   `remote_cache_datastore` (string) - The path to the datastore where
    supporting files will be stored during the build on the remote machine. By
    default this is the same as the `remote_datastore` option. This only has an
    effect if `remote_type` is enabled.

-   `remote_cache_directory` (string) - The path where the ISO and/or floppy
    files will be stored during the build on the remote machine. The path is
    relative to the `remote_cache_datastore` on the remote machine. By default
    this is "packer\_cache". This only has an effect if `remote_type`
    is enabled.

-   `remote_datastore` (string) - The path to the datastore where the resulting
    VM will be stored when it is built on the remote machine. By default this
    is "datastore1". This only has an effect if `remote_type` is enabled.

-   `remote_host` (string) - The host of the remote machine used for access.
    This is only required if `remote_type` is enabled.

-   `remote_password` (string) - The SSH password for the user used to access
    the remote machine. By default this is empty. This only has an effect if
    `remote_type` is enabled.

-   `remote_private_key_file` (string) - The path to the PEM encoded private key
    file for the user used to access the remote machine. By default this is empty.
    This only has an effect if `remote_type` is enabled.

-   `remote_type` (string) - The type of remote machine that will be used to
    build this VM rather than a local desktop product. The only value accepted
    for this currently is "esx5". If this is not set, a desktop product will
    be used. By default, this is not set.

-   `remote_username` (string) - The username for the SSH user that will access
    the remote machine. This is required if `remote_type` is enabled.

-   `shutdown_command` (string) - The command to use to gracefully shut down the
    machine once all the provisioning is done. By default this is an empty
    string, which tells Packer to just forcefully shut down the machine.

-   `shutdown_timeout` (string) - The amount of time to wait after executing the
    `shutdown_command` for the virtual machine to actually shut down. If it
    doesn't shut down in this time, it is an error. By default, the timeout is
    "5m", or five minutes.

-   `skip_compaction` (boolean) - VMware-created disks are defragmented and
    compacted at the end of the build process using `vmware-vdiskmanager`. In
    certain rare cases, this might actually end up making the resulting disks
    slightly larger. If you find this to be the case, you can disable compaction
    using this configuration value.  Defaults to `false`.

-   `keep_registered` (boolean) - Set this to `true` if you would like to keep
    the VM registered with the remote ESXi server. This is convenient if you
    use packer to provision VMs on ESXi and don't want to use ovftool to
    deploy the resulting artifact (VMX or OVA or whatever you used as `format`).
    Defaults to `false`.

-   `tools_upload_flavor` (string) - The flavor of the VMware Tools ISO to
    upload into the VM. Valid values are "darwin", "linux", and "windows". By
    default, this is empty, which means VMware tools won't be uploaded.

-   `tools_upload_path` (string) - The path in the VM to upload the
    VMware tools. This only takes effect if `tools_upload_flavor` is non-empty.
    This is a [configuration
    template](/docs/templates/configuration-templates.html) that has a single
    valid variable: `Flavor`, which will be the value of `tools_upload_flavor`.
    By default the upload path is set to `{{.Flavor}}.iso`. This setting is not
    used when `remote_type` is "esx5".

-   `version` (string) - The [vmx hardware
    version](http://kb.vmware.com/selfservice/microsites/search.do?language=en_US&cmd=displayKC&externalId=1003746)
    for the new virtual machine. Only the default value has been tested, any
    other value is experimental. Default value is '9'.

-   `vm_name` (string) - This is the name of the VMX file for the new virtual
    machine, without the file extension. By default this is "packer-BUILDNAME",
    where "BUILDNAME" is the name of the build.

-   `vmdk_name` (string) - The filename of the virtual disk that'll be created,
    without the extension. This defaults to "packer".

-   `vmx_data` (object of key/value strings) - Arbitrary key/values to enter
    into the virtual machine VMX file. This is for advanced users who want to
    set properties such as memory, CPU, etc.

-   `vmx_data_post` (object of key/value strings) - Identical to `vmx_data`,
    except that it is run after the virtual machine is shutdown, and before the
    virtual machine is exported.

-   `vmx_template_path` (string) - Path to a [configuration
    template](/docs/templates/configuration-templates.html) that defines the
    contents of the virtual machine VMX file for VMware. This is for **advanced
    users only** as this can render the virtual machine non-functional. See
    below for more information. For basic VMX modifications, try
    `vmx_data` first.

-   `vnc_bind_address` (string / IP address) - The IP address that should be binded
     to for VNC. By default packer will use 127.0.0.1 for this. If you wish to bind
     to all interfaces use 0.0.0.0

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
machine, simulating a human actually typing the keyboard. There are a set of
special keys available. If these are in your boot command, they will be replaced
by the proper key:

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

-   `<wait>` `<wait5>` `<wait10>` - Adds a 1, 5 or 10 second pause before
    sending any additional keys. This is useful if you have to generally wait
    for the UI to update before typing more.

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

## VMX Template

The heart of a VMware machine is the "vmx" file. This contains all the virtual
hardware metadata necessary for the VM to function. Packer by default uses a
[safe, flexible VMX
file](https://github.com/mitchellh/packer/blob/20541a7eda085aa5cf35bfed5069592ca49d106e/builder/vmware/step_create_vmx.go#L84).
But for advanced users, this template can be customized. This allows Packer to
build virtual machines of effectively any guest operating system type.

\~&gt; **This is an advanced feature.** Modifying the VMX template can easily
cause your virtual machine to not boot properly. Please only modify the template
if you know what you're doing.

Within the template, a handful of variables are available so that your template
can continue working with the rest of the Packer machinery. Using these
variables isn't required, however.

-   `Name` - The name of the virtual machine.
-   `GuestOS` - The VMware-valid guest OS type.
-   `DiskName` - The filename (without the suffix) of the main virtual disk.
-   `ISOPath` - The path to the ISO to use for the OS installation.
-   `Version` - The Hardware version VMWare will execute this vm under. Also
    known as the `virtualhw.version`.

## Building on a Remote vSphere Hypervisor

In addition to using the desktop products of VMware locally to build virtual
machines, Packer can use a remote VMware Hypervisor to build the virtual
machine.

-&gt; **Note:** Packer supports ESXi 5.1 and above.

Before using a remote vSphere Hypervisor, you need to enable GuestIPHack by
running the following command:

``` {.text}
esxcli system settings advanced set -o /Net/GuestIPHack -i 1
```

When using a remote VMware Hypervisor, the builder still downloads the ISO and
various files locally, and uploads these to the remote machine. Packer currently
uses SSH to communicate to the ESXi machine rather than the vSphere API. At some
point, the vSphere API may be used.

Packer also requires VNC to issue boot commands during a build, which may be
disabled on some remote VMware Hypervisors. Please consult the appropriate
documentation on how to update VMware Hypervisor's firewall to allow these
connections.

To use a remote VMware vSphere Hypervisor to build your virtual machine, fill in
the required `remote_*` configurations:

-   `remote_type` - This must be set to "esx5".

-   `remote_host` - The host of the remote machine.

Additionally, there are some optional configurations that you'll likely have to
modify as well:

-   `remote_port` - The SSH port of the remote machine

-   `remote_datastore` - The path to the datastore where the VM will be stored
    on the ESXi machine.

-   `remote_cache_datastore` - The path to the datastore where supporting files
    will be stored during the build on the remote machine.

-   `remote_cache_directory` - The path where the ISO and/or floppy files will
    be stored during the build on the remote machine. The path is relative to
    the `remote_cache_datastore` on the remote machine.

-   `remote_username` - The SSH username used to access the remote machine.

-   `remote_password` - The SSH password for access to the remote machine.

-   `remote_private_key_file` - The SSH key for access to the remote machine.

-   `format` (string) - Either "ovf", "ova" or "vmx", this specifies the output
    format of the exported virtual machine. This defaults to "ovf".
    Before using this option, you need to install `ovftool`.

### Using a Floppy for Linux kickstart file or preseed

Depending on your network configuration, it may be difficult to use packer's
built-in HTTP server with ESXi. Instead, you can provide a kickstart or preseed
file by attaching a floppy disk. An example below, based on RHEL:

``` {.javascript}
{
  "builders": [
    {
      "type":"vmware-iso",
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
      "type":"vmware-iso",
      "floppy_files": [
        "folder/ks.cfg"
      ],
      "boot_command": "<tab> inst.text inst.ks=hd:fd0:/ks.cfg <enter><wait>"
    }
  ]
}
```
