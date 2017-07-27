---
description: |
    This Parallels builder is able to create Parallels Desktop for Mac virtual
    machines and export them in the PVM format, starting from an existing PVM
    (exported virtual machine image).
layout: docs
page_title: 'Parallels PVM - Builders'
sidebar_current: 'docs-builders-parallels-pvm'
---

# Parallels Builder (from a PVM)

Type: `parallels-pvm`

This Parallels builder is able to create [Parallels Desktop for
Mac](https://www.parallels.com/products/desktop/) virtual machines and export
them in the PVM format, starting from an existing PVM (exported virtual machine
image).

The builder builds a virtual machine by importing an existing PVM file. It then
boots this image, runs provisioners on this new VM, and exports that VM to
create the image. The imported machine is deleted prior to finishing the build.

## Basic Example

Here is a basic example. This example is functional if you have an PVM matching
the settings here.

``` json
{
  "type": "parallels-pvm",
  "parallels_tools_flavor": "lin",
  "source_path": "source.pvm",
  "ssh_username": "packer",
  "ssh_password": "packer",
  "ssh_wait_timeout": "30s",
  "shutdown_command": "echo 'packer' | sudo -S shutdown -P now"
}
```

It is important to add a `shutdown_command`. By default Packer halts the virtual
machine and the file system may not be sync'd. Thus, changes made in a
provisioner might not be saved.

## Configuration Reference

There are many configuration options available for the Parallels builder. They
are organized below into two categories: required and optional. Within each
category, the available options are alphabetized and described.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `parallels_tools_flavor` (string) - The flavor of the Parallels Tools ISO to
    install into the VM. Valid values are "win", "lin", "mac", "os2"
    and "other". This can be omitted only if `parallels_tools_mode`
    is "disable".

-   `source_path` (string) - The path to a PVM directory that acts as the source
    of this build.

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

-   `output_directory` (string) - This is the path to the directory where the
    resulting virtual machine will be created. This may be relative or absolute.
    If relative, the path is relative to the working directory when `packer`
    is executed. This directory must not exist or be empty prior to running
    the builder. By default this is "output-BUILDNAME" where "BUILDNAME" is the
    name of the build.

-   `parallels_tools_guest_path` (string) - The path in the VM to upload
    Parallels Tools. This only takes effect if `parallels_tools_mode`
    is "upload". This is a [configuration
    template](/docs/templates/engine.html) that has a single
    valid variable: `Flavor`, which will be the value of
    `parallels_tools_flavor`. By default this is "prl-tools-{{.Flavor}}.iso"
    which should upload into the login directory of the user.

-   `parallels_tools_mode` (string) - The method by which Parallels Tools are
    made available to the guest for installation. Valid options are "upload",
    "attach", or "disable". If the mode is "attach" the Parallels Tools ISO will
    be attached as a CD device to the virtual machine. If the mode is "upload"
    the Parallels Tools ISO will be uploaded to the path specified by
    `parallels_tools_guest_path`. The default value is "upload".

-   `prlctl` (array of array of strings) - Custom `prlctl` commands to execute
    in order to further customize the virtual machine being created. The value
    of this is an array of commands to execute. The commands are executed in the
    order defined in the template. For each command, the command is defined
    itself as an array of strings, where each string represents a single
    argument on the command-line to `prlctl` (but excluding `prlctl` itself).
    Each arg is treated as a [configuration
    template](/docs/templates/engine.html), where the `Name`
    variable is replaced with the VM name. More details on how to use `prlctl`
    are below.

-   `prlctl_post` (array of array of strings) - Identical to `prlctl`, except
    that it is run after the virtual machine is shutdown, and before the virtual
    machine is exported.

-   `prlctl_version_file` (string) - The path within the virtual machine to
    upload a file that contains the `prlctl` version that was used to create
    the machine. This information can be useful for provisioning. By default
    this is ".prlctl\_version", which will generally upload it into the
    home directory.

-   `reassign_mac` (boolean) - If this is "false" the MAC address of the first
    NIC will reused when imported else a new MAC address will be generated
    by Parallels. Defaults to "false".

-   `shutdown_command` (string) - The command to use to gracefully shut down the
    machine once all the provisioning is done. By default this is an empty
    string, which tells Packer to just forcefully shut down the machine.

-   `shutdown_timeout` (string) - The amount of time to wait after executing the
    `shutdown_command` for the virtual machine to actually shut down. If it
    doesn't shut down in this time, it is an error. By default, the timeout is
    "5m", or five minutes.

-   `skip_compaction` (boolean) - Virtual disk image is compacted at the end of
    the build process using `prl_disk_tool` utility. In certain rare cases, this
    might corrupt the resulting disk image. If you find this to be the case,
    you can disable compaction using this configuration value.

-   `vm_name` (string) - This is the name of the virtual machine when it is
    imported as well as the name of the PVM directory when the virtual machine
    is exported. By default this is "packer-BUILDNAME", where "BUILDNAME" is the
    name of the build.

## Parallels Tools

After the virtual machine is up and the operating system is installed, Packer
uploads the Parallels Tools into the virtual machine. The path where they are
uploaded is controllable by `parallels_tools_path`, and defaults to
"prl-tools.iso". Without an absolute path, it is uploaded to the home directory
of the SSH user. Parallels Tools ISO's can be found in: "/Applications/Parallels
Desktop.app/Contents/Resources/Tools/"

## Boot Command

The `boot_command` specifies the keys to type when the virtual machine is first
booted. This command is typed after `boot_wait`.

As documented above, the `boot_command` is an array of strings. The strings are
all typed in sequence. It is an array only to improve readability within the
template.

The boot command is "typed" character for character (using the Parallels
Virtualization SDK, see [Parallels Builder](/docs/builders/parallels.html))
simulating a human actually typing the keyboard. There are a set of special keys
available. If these are in your boot command, they will be replaced by the
proper key:

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

-   `<leftCtrlOn>` `<rightCtrlOn>` - Simulates pressing and holding the ctrl key.

-   `<leftShiftOn>` `<rightShiftOn>` - Simulates pressing and holding the shift key.

-   `<leftAltOff>` `<rightAltOff>` - Simulates releasing a held alt key.

-   `<leftCtrlOff>` `<rightCtrlOff>` - Simulates releasing a held ctrl key.

-   `<leftShiftOff>` `<rightShiftOff>` - Simulates releasing a held shift key.

-   `<wait>` `<wait5>` `<wait10>` - Adds a 1, 5 or 10 second pause before
    sending any additional keys. This is useful if you have to generally wait
    for the UI to update before typing more.

In addition to the special keys, each command to type is treated as a
[template engine](/docs/templates/engine.html). The
available variables are:

## prlctl Commands

In order to perform extra customization of the virtual machine, a template can
define extra calls to `prlctl` to perform.
[prlctl](http://download.parallels.com/desktop/v9/ga/docs/en_US/Parallels%20Command%20Line%20Reference%20Guide.pdf)
is the command-line interface to Parallels Desktop. It can be used to configure
the virtual machine, such as set RAM, CPUs, etc.

Extra `prlctl` commands are defined in the template in the `prlctl` section. An
example is shown below that sets the memory and number of CPUs within the
virtual machine:

``` json
{
  "prlctl": [
    ["set", "{{.Name}}", "--memsize", "1024"],
    ["set", "{{.Name}}", "--cpus", "2"]
  ]
}
```

The value of `prlctl` is an array of commands to execute. These commands are
executed in the order defined. So in the above example, the memory will be set
followed by the CPUs.

Each command itself is an array of strings, where each string is an argument to
`prlctl`. Each argument is treated as a [configuration
template](/docs/templates/engine.html). The only available
variable is `Name` which is replaced with the unique name of the VM, which is
required for many `prlctl` calls.
