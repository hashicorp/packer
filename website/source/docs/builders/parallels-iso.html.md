---
description: |
    The Parallels Packer builder is able to create Parallels Desktop for Mac virtual
    machines and export them in the PVM format, starting from an ISO image.
layout: docs
page_title: 'Parallels Builder (from an ISO)'
...

# Parallels Builder (from an ISO)

Type: `parallels-iso`

The Parallels Packer builder is able to create [Parallels Desktop for
Mac](https://www.parallels.com/products/desktop/) virtual machines and export
them in the PVM format, starting from an ISO image.

The builder builds a virtual machine by creating a new virtual machine from
scratch, booting it, installing an OS, provisioning software within the OS, then
shutting it down. The result of the Parallels builder is a directory containing
all the files necessary to run the virtual machine portably.

## Basic Example

Here is a basic example. This example is not functional. It will start the OS
installer but then fail because we don't provide the preseed file for Ubuntu to
self-install. Still, the example serves to show the basic configuration:

``` {.javascript}
{
  "type": "parallels-iso",
  "guest_os_type": "ubuntu",
  "iso_url": "http://releases.ubuntu.com/12.04/ubuntu-12.04.3-server-amd64.iso",
  "iso_checksum": "2cbe868812a871242cdcdd8f2fd6feb9",
  "iso_checksum_type": "md5",
  "parallels_tools_flavor": "lin",
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

-   `parallels_tools_flavor` (string) - The flavor of the Parallels Tools ISO to
    install into the VM. Valid values are "win", "lin", "mac", "os2"
    and "other". This can be omitted only if `parallels_tools_mode`
    is "disable".

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

-   `disk_size` (integer) - The size, in megabytes, of the hard disk to create
    for the VM. By default, this is 40000 (about 40 GB).

-   `floppy_files` (array of strings) - A list of files to place onto a floppy
    disk that is attached when the VM is booted. This is most useful for
    unattended Windows installs, which look for an `Autounattend.xml` file on
    removable media. By default, no floppy will be attached. All files listed in
    this setting get placed into the root directory of the floppy and the floppy
    is attached as the first floppy device. Currently, no support exists for
    creating sub-directories on the floppy. Wildcard characters (\*, ?,
    and \[\]) are allowed. Directory names are also allowed, which will add all
    the files found in the directory to the floppy.

-   `guest_os_type` (string) - The guest OS type being installed. By default
    this is "other", but you can get *dramatic* performance improvements by
    setting this to the proper value. To view all available values for this run
    `prlctl create x --distribution list`. Setting the correct value hints to
    Parallels Desktop how to optimize the virtual hardware to work best with
    that operating system.

-   `hard_drive_interface` (string) - The type of controller that the hard
    drives are attached to, defaults to "sata". Valid options are "sata", "ide",
    and "scsi".

-   `host_interfaces` (array of strings) - A list of which interfaces on the
    host should be searched for a IP address. The first IP address found on one
    of these will be used as `{{ .HTTPIP }}` in the `boot_command`. Defaults to
    \["en0", "en1", "en2", "en3", "en4", "en5", "en6", "en7", "en8", "en9",
    "ppp0", "ppp1", "ppp2"\].

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

-   `parallels_tools_guest_path` (string) - The path in the virtual machine to
    upload Parallels Tools. This only takes effect if `parallels_tools_mode`
    is "upload". This is a [configuration
    template](/docs/templates/configuration-templates.html) that has a single
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
    template](/docs/templates/configuration-templates.html), where the `Name`
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

-   `vm_name` (string) - This is the name of the PVM directory for the new
    virtual machine, without the file extension. By default this is
    "packer-BUILDNAME", where "BUILDNAME" is the name of the build.

## Boot Command

The `boot_command` configuration is very important: it specifies the keys to
type when the virtual machine is first booted in order to start the OS
installer. This command is typed after `boot_wait`, which gives the virtual
machine some time to actually load the ISO.

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
  "initrd=/install/initrd.gz -- <enter>;"
]
```

## prlctl Commands

In order to perform extra customization of the virtual machine, a template can
define extra calls to `prlctl` to perform.
[prlctl](http://download.parallels.com/desktop/v9/ga/docs/en_US/Parallels%20Command%20Line%20Reference%20Guide.pdf)
is the command-line interface to Parallels Desktop. It can be used to configure
the virtual machine, such as set RAM, CPUs, etc.

Extra `prlctl` commands are defined in the template in the `prlctl` section. An
example is shown below that sets the memory and number of CPUs within the
virtual machine:

``` {.javascript}
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
template](/docs/templates/configuration-templates.html). The only available
variable is `Name` which is replaced with the unique name of the VM, which is
required for many `prlctl` calls.
