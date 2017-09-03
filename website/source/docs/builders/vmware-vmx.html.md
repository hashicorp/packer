---
description: |
    This VMware Packer builder is able to create VMware virtual machines from an
    existing VMware virtual machine (a VMX file). It currently supports building
    virtual machines on hosts running VMware Fusion Professional for OS X, VMware
    Workstation for Linux and Windows, and VMware Player on Linux.
layout: docs
page_title: 'VMware VMX - Builders'
sidebar_current: 'docs-builders-vmware-vmx'
---

# VMware Builder (from VMX)

Type: `vmware-vmx`

This VMware Packer builder is able to create VMware virtual machines from an
existing VMware virtual machine (a VMX file). It currently supports building
virtual machines on hosts running [VMware Fusion
Professional](https://www.vmware.com/products/fusion-professional/) for OS X,
[VMware Workstation](https://www.vmware.com/products/workstation/overview.html)
for Linux and Windows, and [VMware
Player](https://www.vmware.com/products/player/) on Linux.

The builder builds a virtual machine by cloning the VMX file using the clone
capabilities introduced in VMware Fusion Professional 6, Workstation 10, and
Player 6. After cloning the VM, it provisions software within the new machine,
shuts it down, and compacts the disks. The resulting folder contains a new
VMware virtual machine.

## Basic Example

Here is an example. This example is fully functional as long as the source path
points to a real VMX file with the proper settings:

``` json
{
  "type": "vmware-vmx",
  "source_path": "/path/to/a/vm.vmx",
  "ssh_username": "root",
  "ssh_password": "root",
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

-   `source_path` (string) - Path to the source VMX file to clone.

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

-   `fusion_app_path` (string) - Path to "VMware Fusion.app". By default this is
    "/Applications/VMware Fusion.app" but this setting allows you to
    customize this.

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

-   `output_directory` (string) - This is the path to the directory where the
    resulting virtual machine will be created. This may be relative or absolute.
    If relative, the path is relative to the working directory when `packer`
    is executed. This directory must not exist or be empty prior to running
    the builder. By default this is "output-BUILDNAME" where "BUILDNAME" is the
    name of the build.

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
    "5m", or five minutes.

-   `skip_compaction` (boolean) - VMware-created disks are defragmented and
    compacted at the end of the build process using `vmware-vdiskmanager`. In
    certain rare cases, this might actually end up making the resulting disks
    slightly larger. If you find this to be the case, you can disable compaction
    using this configuration value.

-   `tools_upload_flavor` (string) - The flavor of the VMware Tools ISO to
    upload into the VM. Valid values are "darwin", "linux", and "windows". By
    default, this is empty, which means VMware tools won't be uploaded.

-   `tools_upload_path` (string) - The path in the VM to upload the
    VMware tools. This only takes effect if `tools_upload_flavor` is non-empty.
    This is a [configuration
    template](/docs/templates/engine.html) that has a single
    valid variable: `Flavor`, which will be the value of `tools_upload_flavor`.
    By default the upload path is set to `{{.Flavor}}.iso`.

-   `vm_name` (string) - This is the name of the VMX file for the new virtual
    machine, without the file extension. By default this is "packer-BUILDNAME",
    where "BUILDNAME" is the name of the build.

-   `vmx_data` (object of key/value strings) - Arbitrary key/values to enter
    into the virtual machine VMX file. This is for advanced users who want to
    set properties such as memory, CPU, etc.

-   `vmx_data_post` (object of key/value strings) - Identical to `vmx_data`,
    except that it is run after the virtual machine is shutdown, and before the
    virtual machine is exported.

-   `vmx_remove_ethernet_interfaces` (boolean) - Remove all ethernet interfaces from
    the VMX file after building. This is for advanced users who understand the
    ramifications, but is useful for building Vagrant boxes since Vagrant will
    create ethernet interfaces when provisioning a box.

-   `vnc_bind_address` (string / IP address) - The IP address that should be binded
    to for VNC. By default packer will use 127.0.0.1 for this.

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
installer. This command is typed after `boot_wait`.

As documented above, the `boot_command` is an array of strings. The strings are
all typed in sequence. It is an array only to improve readability within the
template.

The boot command is "typed" character for character over a VNC connection to the
machine, simulating a human actually typing the keyboard.

-&gt; Keystrokes are typed as separate key up/down events over VNC with a
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

-   `<leftAlt>` `<rightAlt>` - Simulates pressing the alt key.

-   `<leftCtrl>` `<rightCtrl>` - Simulates pressing the ctrl key.

-   `<leftShift>` `<rightShift>` - Simulates pressing the shift key.

-   `<leftAltOn>` `<rightAltOn>` - Simulates pressing and holding the alt key.

-   `<leftCtrlOn>` `<rightCtrlOn>` - Simulates pressing and holding the ctrl
    key.

-   `<leftShiftOn>` `<rightShiftOn>` - Simulates pressing and holding the
    shift key.

-   `<leftAltOff>` `<rightAltOff>` - Simulates releasing a held alt key.

-   `<leftCtrlOff>` `<rightCtrlOff>` - Simulates releasing a held ctrl key.

-   `<leftShiftOff>` `<rightShiftOff>` - Simulates releasing a held shift key.

-   `<wait>` `<wait5>` `<wait10>` - Adds a 1, 5 or 10 second pause before
    sending any additional keys. This is useful if you have to generally wait
    for the UI to update before typing more.

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
