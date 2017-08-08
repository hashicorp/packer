---
description: |
    The VirtualBox Packer builder is able to create VirtualBox virtual machines
    and export them in the OVF format, starting from an ISO image.
layout: docs
page_title: 'VirtualBox ISO - Builders'
sidebar_current: 'docs-builders-virtualbox-iso'
---

# VirtualBox Builder (from an ISO)

Type: `virtualbox-iso`

The VirtualBox Packer builder is able to create
[VirtualBox](https://www.virtualbox.org/) virtual machines and export them in
the OVF format, starting from an ISO image.

The builder builds a virtual machine by creating a new virtual machine from
scratch, booting it, installing an OS, provisioning software within the OS, then
shutting it down. The result of the VirtualBox builder is a directory containing
all the files necessary to run the virtual machine portably.

## Basic Example

Here is a basic example. This example is not functional. It will start the OS
installer but then fail because we don't provide the preseed file for Ubuntu to
self-install. Still, the example serves to show the basic configuration:

``` json
{
  "type": "virtualbox-iso",
  "guest_os_type": "Ubuntu_64",
  "iso_url": "http://releases.ubuntu.com/12.04/ubuntu-12.04.5-server-amd64.iso",
  "iso_checksum": "769474248a3897f4865817446f9a4a53",
  "iso_checksum_type": "md5",
  "ssh_username": "packer",
  "ssh_password": "packer",
  "shutdown_command": "echo 'packer' | sudo -S shutdown -P now"
}
```

It is important to add a `shutdown_command`. By default Packer halts the virtual
machine and the file system may not be sync'd. Thus, changes made in a
provisioner might not be saved.

## Configuration Reference

There are many configuration options available for the VirtualBox builder. They
are organized below into two categories: required and optional. Within each
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

-   `disk_size` (integer) - The size, in megabytes, of the hard disk to create
    for the VM. By default, this is 40000 (about 40 GB).

-   `export_opts` (array of strings) - Additional options to pass to the
    [VBoxManage
    export](https://www.virtualbox.org/manual/ch08.html#vboxmanage-export). This
    can be useful for passing product information to include in the resulting
    appliance file. Packer JSON configuration file example:

    ``` json
    {
      "type": "virtualbox-iso",
      "export_opts":
      [
        "--manifest",
        "--vsys", "0",
        "--description", "{{user `vm_description`}}",
        "--version", "{{user `vm_version`}}"
      ],
      "format": "ova",
    }
    ```

    A VirtualBox [VM
    description](https://www.virtualbox.org/manual/ch08.html#idm3756) may
    contain arbitrary strings; the GUI interprets HTML formatting. However, the
    JSON format does not allow arbitrary newlines within a value. Add a
    multi-line description by preparing the string in the shell before the
    packer call like this (shell `>` continuation character snipped for easier
    copy & paste):

    ``` {.shell}

    vm_description='some
    multiline
    description'

    vm_version='0.2.0'

    packer build \
        -var "vm_description=${vm_description}" \
        -var "vm_version=${vm_version}"         \
        "packer_conf.json"
    ```

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

-   `format` (string) - Either "ovf" or "ova", this specifies the output format
    of the exported virtual machine. This defaults to "ovf".

-   `guest_additions_mode` (string) - The method by which guest additions are
    made available to the guest for installation. Valid options are "upload",
    "attach", or "disable". If the mode is "attach" the guest additions ISO will
    be attached as a CD device to the virtual machine. If the mode is "upload"
    the guest additions ISO will be uploaded to the path specified by
    `guest_additions_path`. The default value is "upload". If "disable" is used,
    guest additions won't be downloaded, either.

-   `guest_additions_path` (string) - The path on the guest virtual machine
    where the VirtualBox guest additions ISO will be uploaded. By default this
    is "VBoxGuestAdditions.iso" which should upload into the login directory of
    the user. This is a [configuration
    template](/docs/templates/engine.html) where the `Version`
    variable is replaced with the VirtualBox version.

-   `guest_additions_sha256` (string) - The SHA256 checksum of the guest
    additions ISO that will be uploaded to the guest VM. By default the
    checksums will be downloaded from the VirtualBox website, so this only needs
    to be set if you want to be explicit about the checksum.

-   `guest_additions_url` (string) - The URL to the guest additions ISO
    to upload. This can also be a file URL if the ISO is at a local path. By
    default, the VirtualBox builder will attempt to find the guest additions ISO
    on the local file system. If it is not available locally, the builder will
    download the proper guest additions ISO from the internet.

-   `guest_os_type` (string) - The guest OS type being installed. By default
    this is "other", but you can get *dramatic* performance improvements by
    setting this to the proper value. To view all available values for this run
    `VBoxManage list ostypes`. Setting the correct value hints to VirtualBox how
    to optimize the virtual hardware to work best with that operating system.

-   `hard_drive_interface` (string) - The type of controller that the primary
    hard drive is attached to, defaults to "ide". When set to "sata", the drive
    is attached to an AHCI SATA controller. When set to "scsi", the drive is
    attached to an LsiLogic SCSI controller.

-   `sata_port_count` (integer) - The number of ports available on any SATA
    controller created, defaults to 1. VirtualBox supports up to 30 ports on a
    maxiumum of 1 SATA controller. Increasing this value can be useful if you
    want to attach additional drives.

-   `hard_drive_nonrotational` (boolean) - Forces some guests (i.e. Windows 7+)
    to treat disks as SSDs and stops them from performing disk fragmentation.
    Also set `hard_drive_Discard` to `true` to enable TRIM support.

-   `hard_drive_discard` (boolean) - When this value is set to `true`, a VDI
    image will be shrunk in response to the trim command from the guest OS.
    The size of the cleared area must be at least 1MB. Also set
    `hard_drive_nonrotational` to `true` to enable TRIM support.

-   `headless` (boolean) - Packer defaults to building VirtualBox virtual
    machines by launching a GUI that shows the console of the machine
    being built. When this value is set to `true`, the machine will start without
    a console.

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

-   `iso_interface` (string) - The type of controller that the ISO is attached
    to, defaults to "ide". When set to "sata", the drive is attached to an AHCI
    SATA controller.

-   `iso_target_extension` (string) - The extension of the iso file after
    download. This defaults to "iso".

-   `iso_target_path` (string) - The path where the iso should be saved
    after download. By default will go in the packer cache, with a hash of the
    original filename as its name.

-   `iso_urls` (array of strings) - Multiple URLs for the ISO to download.
    Packer will try these in order. If anything goes wrong attempting to
    download or while downloading a single URL, it will move on to the next. All
    URLs must point to the same file (same checksum). By default this is empty
    and `iso_url` is used. Only one of `iso_url` or `iso_urls` can be specified.

-   `keep_registered` (boolean) - Set this to `true` if you would like to keep
    the VM registered with virtualbox. Defaults to `false`.

-   `output_directory` (string) - This is the path to the directory where the
    resulting virtual machine will be created. This may be relative or absolute.
    If relative, the path is relative to the working directory when `packer`
    is executed. This directory must not exist or be empty prior to running
    the builder. By default this is "output-BUILDNAME" where "BUILDNAME" is the
    name of the build.

-   `post_shutdown_delay` (string) - The amount of time to wait after shutting
    down the virtual machine. If you get the error
    `Error removing floppy controller`, you might need to set this to `5m`
    or so. By default, the delay is `0s`, or disabled.

-   `shutdown_command` (string) - The command to use to gracefully shut down the
    machine once all the provisioning is done. By default this is an empty
    string, which tells Packer to just forcefully shut down the machine unless a
    shutdown command takes place inside script so this may safely be omitted. If
    one or more scripts require a reboot it is suggested to leave this blank
    since reboots may fail and specify the final shutdown command in your
    last script.

-   `shutdown_timeout` (string) - The amount of time to wait after executing the
    `shutdown_command` for the virtual machine to actually shut down. If it
    doesn't shut down in this time, it is an error. By default, the timeout is
    `5m`, or five minutes.

-   `skip_export` (boolean) - Defaults to `false`. When enabled, Packer will
    not export the VM. Useful if the build output is not the resultant image,
    but created inside the VM.

-   `ssh_host_port_min` and `ssh_host_port_max` (integer) - The minimum and
    maximum port to use for the SSH port on the host machine which is forwarded
    to the SSH port on the guest machine. Because Packer often runs in parallel,
    Packer will choose a randomly available port in this range to use as the
    host port. By default this is 2222 to 4444.

-   `ssh_skip_nat_mapping` (boolean) - Defaults to `false`. When enabled, Packer
    does not setup forwarded port mapping for SSH requests and uses `ssh_port`
    on the host to communicate to the virtual machine

-   `vboxmanage` (array of array of strings) - Custom `VBoxManage` commands to
    execute in order to further customize the virtual machine being created. The
    value of this is an array of commands to execute. The commands are executed
    in the order defined in the template. For each command, the command is
    defined itself as an array of strings, where each string represents a single
    argument on the command-line to `VBoxManage` (but excluding
    `VBoxManage` itself). Each arg is treated as a [configuration
    template](/docs/templates/engine.html), where the `Name`
    variable is replaced with the VM name. More details on how to use
    `VBoxManage` are below.

-   `vboxmanage_post` (array of array of strings) - Identical to `vboxmanage`,
    except that it is run after the virtual machine is shutdown, and before the
    virtual machine is exported.

-   `virtualbox_version_file` (string) - The path within the virtual machine to
    upload a file that contains the VirtualBox version that was used to create
    the machine. This information can be useful for provisioning. By default
    this is ".vbox\_version", which will generally be upload it into the
    home directory. Set to an empty string to skip uploading this file, which
    can be useful when using the `none` communicator.

-   `vm_name` (string) - This is the name of the OVF file for the new virtual
    machine, without the file extension. By default this is "packer-BUILDNAME",
    where "BUILDNAME" is the name of the build.

-   `vrdp_bind_address` (string / IP address) - The IP address that should be
    binded to for VRDP. By default packer will use 127.0.0.1 for this. If you
    wish to bind to all interfaces use 0.0.0.0

-   `vrdp_port_min` and `vrdp_port_max` (integer) - The minimum and maximum port
    to use for VRDP access to the virtual machine. Packer uses a randomly chosen
    port in this range that appears available. By default this is 5900 to 6000.
    The minimum and maximum ports are inclusive.

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

-   `<leftAlt>` `<rightAlt>` - Simulates pressing the alt key.

-   `<leftCtrl>` `<rightCtrl>` - Simulates pressing the ctrl key.

-   `<leftShift>` `<rightShift>` - Simulates pressing the shift key.

-   `<leftAltOn>` `<rightAltOn>` - Simulates pressing and holding the alt key.

-   `<leftCtrlOn>` `<rightCtrlOn>` - Simulates pressing and holding the
    ctrl key.

-   `<leftShiftOn>` `<rightShiftOn>` - Simulates pressing and holding the
    shift key.

-   `<leftAltOff>` `<rightAltOff>` - Simulates releasing a held alt key.

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
[template engine](/docs/templates/engine.html). The
available variables are:

-   `HTTPIP` and `HTTPPort` - The IP and port, respectively of an HTTP server
    that is started serving the directory specified by the `http_directory`
    configuration parameter. If `http_directory` isn't specified, these will be
    blank!

Example boot command. This is actually a working boot command used to start an
Ubuntu 12.04 installer:

``` text
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

## Guest Additions

Packer will automatically download the proper guest additions for the version of
VirtualBox that is running and upload those guest additions into the virtual
machine so that provisioners can easily install them.

Packer downloads the guest additions from the official VirtualBox website, and
verifies the file with the official checksums released by VirtualBox.

After the virtual machine is up and the operating system is installed, Packer
uploads the guest additions into the virtual machine. The path where they are
uploaded is controllable by `guest_additions_path`, and defaults to
"VBoxGuestAdditions.iso". Without an absolute path, it is uploaded to the home
directory of the SSH user.

## VBoxManage Commands

In order to perform extra customization of the virtual machine, a template can
define extra calls to `VBoxManage` to perform.
[VBoxManage](https://www.virtualbox.org/manual/ch08.html) is the command-line
interface to VirtualBox where you can completely control VirtualBox. It can be
used to do things such as set RAM, CPUs, etc.

Extra VBoxManage commands are defined in the template in the `vboxmanage`
section. An example is shown below that sets the memory and number of CPUs
within the virtual machine:

``` json
{
  "vboxmanage": [
    ["modifyvm", "{{.Name}}", "--memory", "1024"],
    ["modifyvm", "{{.Name}}", "--cpus", "2"]
  ]
}
```

The value of `vboxmanage` is an array of commands to execute. These commands are
executed in the order defined. So in the above example, the memory will be set
followed by the CPUs.

Each command itself is an array of strings, where each string is an argument to
`VBoxManage`. Each argument is treated as a [configuration
template](/docs/templates/engine.html). The only available
variable is `Name` which is replaced with the unique name of the VM, which is
required for many VBoxManage calls.
