---
description: |
    The Qemu Packer builder is able to create KVM and Xen virtual machine images.
    Support for Xen is experimental at this time.
layout: docs
page_title: QEMU Builder
...

# QEMU Builder

Type: `qemu`

The Qemu Packer builder is able to create [KVM](http://www.linux-kvm.org) and
[Xen](http://www.xenproject.org) virtual machine images. Support for Xen is
experimental at this time.

The builder builds a virtual machine by creating a new virtual machine from
scratch, booting it, installing an OS, rebooting the machine with the boot media
as the virtual hard drive, provisioning software within the OS, then shutting it
down. The result of the Qemu builder is a directory containing the image file
necessary to run the virtual machine on KVM or Xen.

## Basic Example

Here is a basic example. This example is functional so long as you fixup paths
to files, URLS for ISOs and checksums.

``` {.javascript}
{
  "builders":
  [
    {
      "type": "qemu",
      "iso_url": "http://mirror.raystedman.net/centos/6/isos/x86_64/CentOS-6.5-x86_64-minimal.iso",
      "iso_checksum": "0d9dc37b5dd4befa1c440d2174e88a87",
      "iso_checksum_type": "md5",
      "output_directory": "output_centos_tdhtest",
      "shutdown_command": "shutdown -P now",
      "disk_size": 5000,
      "format": "qcow2",
      "headless": false,
      "accelerator": "kvm",
      "http_directory": "httpdir",
      "http_port_min": 10082,
      "http_port_max": 10089,
      "ssh_host_port_min": 2222,
      "ssh_host_port_max": 2229,
      "ssh_username": "root",
      "ssh_password": "s0m3password",
      "ssh_port": 22,
      "ssh_wait_timeout": "30s",
      "vm_name": "tdhtest",
      "net_device": "virtio-net",
      "disk_interface": "virtio",
      "boot_wait": "5s",
      "boot_command":
      [
        "<tab> text ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/centos6-ks.cfg<enter><wait>"
      ]
    }
  ]
}
```

A working CentOS 6.x kickstart file can be found [at this
URL](https://gist.github.com/mitchellh/7328271/#file-centos6-ks-cfg), adapted
from an unknown source. Place this file in the http directory with the proper
name. For the example above, it should go into "httpdir" with a name of
"centos6-ks.cfg".

## Configuration Reference

There are many configuration options available for the Qemu builder. They are
organized below into two categories: required and optional. Within each
category, the available options are alphabetized and described.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

Note that you will need to set `"headless": true` if you are running Packer
on a Linux server without X11; or if you are connected via ssh to a remote
Linux server and have not enabled X11 forwarding (`ssh -X`).

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

-   `accelerator` (string) - The accelerator type to use when running the VM.
    This may be `none`, `kvm`, `tcg`, or `xen`. The appropriate software must
    already been installed on your build machine to use the accelerator you
    specified. When no accelerator is specified, Packer will try to use `kvm`
    if it is available but will default to `tcg` otherwise.

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

-   `disk_cache` (string) - The cache mode to use for disk. Allowed values
    include any of "writethrough", "writeback", "none", "unsafe"
    or "directsync". By default, this is set to "writeback".

-   `disk_compression` (boolean) - Apply compression to the QCOW2 disk file
    using `qemu-img convert`. Defaults to `false`.

-   `disk_discard` (string) - The discard mode to use for disk. Allowed values
    include any of "unmap" or "ignore". By default, this is set to "ignore".

-   `disk_image` (boolean) - Packer defaults to building from an ISO file, this
    parameter controls whether the ISO URL supplied is actually a bootable
    QEMU image. When this value is set to true, the machine will clone the
    source, resize it according to `disk_size` and boot the image.

-   `disk_interface` (string) - The interface to use for the disk. Allowed
    values include any of "ide", "scsi", "virtio" or "virtio-scsi". Note also
    that any boot commands or kickstart type scripts must have proper
    adjustments for resulting device names. The Qemu builder uses "virtio" by
    default.

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

-   `format` (string) - Either "qcow2" or "raw", this specifies the output
    format of the virtual machine image. This defaults to `qcow2`.

-   `headless` (boolean) - Packer defaults to building QEMU virtual machines by
    launching a GUI that shows the console of the machine being built. When this
    value is set to true, the machine will start without a console.

    You can still see the console if you make a note of the VNC display
    number chosen, and then connect using `vncviewer -Shared <host>:<display>`

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

-   `iso_skip_cache` (boolean) - Use iso from provided url. Qemu must support
    curl block device. This defaults to `false`.

-   `iso_target_path` (string) - The path where the iso should be saved after
    download. By default will go in the packer cache, with a hash of the
    original filename as its name.

-   `iso_urls` (array of strings) - Multiple URLs for the ISO to download.
    Packer will try these in order. If anything goes wrong attempting to
    download or while downloading a single URL, it will move on to the next. All
    URLs must point to the same file (same checksum). By default this is empty
    and `iso_url` is used. Only one of `iso_url` or `iso_urls` can be specified.

-   `machine_type` (string) - The type of machine emulation to use. Run your
    qemu binary with the flags `-machine help` to list available types for
    your system. This defaults to "pc".

-   `net_device` (string) - The driver to use for the network interface. Allowed
    values "ne2k\_pci", "i82551", "i82557b", "i82559er", "rtl8139", "e1000",
    "pcnet", "virtio", "virtio-net", "virtio-net-pci", "usb-net", "i82559a",
    "i82559b", "i82559c", "i82550", "i82562", "i82557a", "i82557c", "i82801",
    "vmxnet3", "i82558a" or "i82558b".  The Qemu builder uses "virtio-net" by
    default.

-   `output_directory` (string) - This is the path to the directory where the
    resulting virtual machine will be created. This may be relative or absolute.
    If relative, the path is relative to the working directory when `packer`
    is executed. This directory must not exist or be empty prior to running
    the builder. By default this is "output-BUILDNAME" where "BUILDNAME" is the
    name of the build.

-   `qemu_binary` (string) - The name of the Qemu binary to look for. This
    defaults to "qemu-system-x86\_64", but may need to be changed for
    some platforms. For example "qemu-kvm", or "qemu-system-i386" may be a
    better choice for some systems.

-   `qemuargs` (array of array of strings) - Allows complete control over the
    qemu command line (though not, at this time, qemu-img). Each array of
    strings makes up a command line switch that overrides matching default
    switch/value pairs. Any value specified as an empty string is ignored. All
    values after the switch are concatenated with no separator.

\~&gt; **Warning:** The qemu command line allows extreme flexibility, so beware
of conflicting arguments causing failures of your run. For instance, using
--no-acpi could break the ability to send power signal type commands (e.g.,
shutdown -P now) to the virtual machine, thus preventing proper shutdown. To see
the defaults, look in the packer.log file and search for the qemu-system-x86
command. The arguments are all printed for review.

The following shows a sample usage:

``` {.javascript}
  // ...
  "qemuargs": [
    [ "-m", "1024M" ],
    [ "--no-acpi", "" ],
    [
       "-netdev",
      "user,id=mynet0,",
      "hostfwd=hostip:hostport-guestip:guestport",
      ""
    ],
    [ "-device", "virtio-net,netdev=mynet0" ]
  ]
  // ...
```

would produce the following (not including other defaults supplied by the
builder and not otherwise conflicting with the qemuargs):

<pre class="prettyprint">
  qemu-system-x86 -m 1024m --no-acpi -netdev user,id=mynet0,hostfwd=hostip:hostport-guestip:guestport -device virtio-net,netdev=mynet0"
</pre>

You can also use the `SSHHostPort` template variable to produce a packer
template that can be invoked by `make` in parallel:

``` {.javascript}
  // ...
  "qemuargs": [
          [ "-netdev", "user,hostfwd=tcp::{{ .SSHHostPort }}-:22,id=forward"],
          [ "-device", "virtio-net,netdev=forward,id=net0"],
          ...
        ]
  // ...
```
`make -j 3 my-awesome-packer-templates` spawns 3 packer processes, each of which
will bind to their own SSH port as determined by each process. This will also
work with WinRM, just change the port forward in `qemuargs` to map to WinRM's
default port of `5985` or whatever value you have the service set to listen on.

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

-   `skip_compaction` (boolean) - Packer compacts the QCOW2 image using `qemu-img convert`.
    Set this option to `true` to disable compacting. Defaults to `false`.

-   `ssh_host_port_min` and `ssh_host_port_max` (integer) - The minimum and
    maximum port to use for the SSH port on the host machine which is forwarded
    to the SSH port on the guest machine. Because Packer often runs in parallel,
    Packer will choose a randomly available port in this range to use as the
    host port. By default this is 2222 to 4444.

-   `vm_name` (string) - This is the name of the image (QCOW2 or IMG) file for
    the new virtual machine. By default this is "packer-BUILDNAME", where
    `BUILDNAME` is the name of the build. Currently, no file extension will be
    used unless it is specified in this option.

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

-   `<waitXX> ` - Add user defined time.Duration pause before sending any
    additional keys. For example `<wait10m>` or `<wait1m20s>`

In addition to the special keys, each command to type is treated as a
[configuration template](/docs/templates/configuration-templates.html). The
available variables are:

-   `HTTPIP` and `HTTPPort` - The IP and port, respectively of an HTTP server
    that is started serving the directory specified by the `http_directory`
    configuration parameter. If `http_directory` isn't specified, these will be
    blank!

Example boot command. This is actually a working boot command used to start an
CentOS 6.4 installer:

``` {.javascript}
"boot_command":
[
  "<tab><wait>",
  " ks=http://10.0.2.2:{{ .HTTPPort }}/centos6-ks.cfg<enter>"
]
```
