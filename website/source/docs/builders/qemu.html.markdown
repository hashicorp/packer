---
layout: "docs"
---

# QEMU Builder

Type: `qemu`

The Qemu builder is able to create [KVM](http://www.linux-kvm.org)
and [Xen](http://www.xenproject.org) virtual machine images. Support
for Xen is experimental at this time.

The builder builds a virtual machine by creating a new virtual machine
from scratch, booting it, installing an OS, rebooting the machine with the
boot media as the virtual hard drive, provisioning software within
the OS, then shutting it down. The result of the Qemu builder is a directory
containing the image file necessary to run the virtual machine on KVM or Xen.

## Basic Example

Here is a basic example. This example is functional so long as you fixup
paths to files, URLS for ISOs and checksums.

<pre class="prettyprint">
{
  "builders":
  [
    {
      "type": "qemu",
      "iso_url": "http://mirror.raystedman.net/centos/6/isos/x86_64/CentOS-6.4-x86_64-minimal.iso",
      "iso_checksum": "4a5fa01c81cc300f4729136e28ebe600",
      "iso_checksum_type": "md5",
      "output_directory": "output_centos_tdhtest",
      "ssh_wait_timeout": "30s",
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
      "ssh_wait_timeout": "90m",
      "vm_name": "tdhtest",
      "net_device": "virtio-net",
      "disk_interface": "virtio",
      "boot_command":
      [
        "<tab><wait>",
        " ks=http://10.0.2.2:{{ .HTTPPort }}/centos6-ks.cfg<enter>"
      ]
    }
  ]
}
</pre>

A working CentOS 6.x kickstart file can be found
[at this URL](https://gist.github.com/mitchellh/7328271/raw/c91e0c4fa19c171a40b016c6c8f251f90d2ad0ba/centos6-ks.cfg), adapted from an unknown source.
Place this file in the http directory with the proper name. For the
example above, it should go into "httpdir" with a name of "centos6-ks.cfg".

## Configuration Reference

There are many configuration options available for the Qemu builder.
They are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

### Required:

* `iso_checksum` (string) - The checksum for the OS ISO file. Because ISO
  files are so large, this is required and Packer will verify it prior
  to booting a virtual machine with the ISO attached. The type of the
  checksum is specified with `iso_checksum_type`, documented below.

* `iso_checksum_type` (string) - The type of the checksum specified in
  `iso_checksum`. Valid values are "md5", "sha1", "sha256", or "sha512" currently.

* `iso_url` (string) - A URL to the ISO containing the installation image.
  This URL can be either an HTTP URL or a file URL (or path to a file).
  If this is an HTTP URL, Packer will download it and cache it between
  runs.

* `ssh_username` (string) - The username to use to SSH into the machine
  once the OS is installed.

### Optional:

* `accelerator` (string) - The accelerator type to use when running the VM.
  This may have a value of either "kvm" or "xen" and you must have that
  support in on the machine on which you run the builder.

* `boot_command` (array of strings) - This is an array of commands to type
  when the virtual machine is first booted. The goal of these commands should
  be to type just enough to initialize the operating system installer. Special
  keys can be typed as well, and are covered in the section below on the boot
  command. If this is not specified, it is assumed the installer will start
  itself.

* `boot_wait` (string) - The time to wait after booting the initial virtual
  machine before typing the `boot_command`. The value of this should be
  a duration. Examples are "5s" and "1m30s" which will cause Packer to wait
  five seconds and one minute 30 seconds, respectively. If this isn't specified,
  the default is 10 seconds.

* `disk_size` (integer) - The size, in megabytes, of the hard disk to create
  for the VM. By default, this is 40000 (about 40 GB).

* `disk_interface` (string) - The interface to use for the disk. Allowed
  values include any of "ide," "scsi" or "virtio." Note also that any boot
  commands or kickstart type scripts must have proper adjustments for
  resulting device names. The Qemu builder uses "virtio" by default.

* `floppy_files` (array of strings) - A list of files to place onto a floppy
  disk that is attached when the VM is booted. This is most useful
  for unattended Windows installs, which look for an `Autounattend.xml` file
  on removable media. By default, no floppy will be attached. All files
  listed in this setting get placed into the root directory of the floppy
  and the floppy is attached as the first floppy device. Currently, no
  support exists for creating sub-directories on the floppy. Wildcard
  characters (*, ?, and []) are allowed. Directory names are also allowed,
  which will add all the files found in the directory to the floppy.

* `format` (string) - Either "qcow2" or "raw", this specifies the output
  format of the virtual machine image. This defaults to "qcow2".

* `headless` (boolean) - Packer defaults to building virtual machines by
  launching a GUI that shows the console of the machine being built.
  When this value is set to true, the machine will start without a console.

* `http_directory` (string) - Path to a directory to serve using an HTTP
  server. The files in this directory will be available over HTTP that will
  be requestable from the virtual machine. This is useful for hosting
  kickstart files and so on. By default this is "", which means no HTTP
  server will be started. The address and port of the HTTP server will be
  available as variables in `boot_command`. This is covered in more detail
  below.

* `http_port_min` and `http_port_max` (integer) - These are the minimum and
  maximum port to use for the HTTP server started to serve the `http_directory`.
  Because Packer often runs in parallel, Packer will choose a randomly available
  port in this range to run the HTTP server. If you want to force the HTTP
  server to be on one port, make this minimum and maximum port the same.
  By default the values are 8000 and 9000, respectively.

* `iso_urls` (array of strings) - Multiple URLs for the ISO to download.
  Packer will try these in order. If anything goes wrong attempting to download
  or while downloading a single URL, it will move on to the next. All URLs
  must point to the same file (same checksum). By default this is empty
  and `iso_url` is used. Only one of `iso_url` or `iso_urls` can be specified.

* `net_device` (string) - The driver to use for the network interface. Allowed
  values "ne2k_pci," "i82551," "i82557b," "i82559er," "rtl8139," "e1000,"
  "pcnet" or "virtio." The Qemu builder uses "virtio" by default.

* `output_directory` (string) - This is the path to the directory where the
  resulting virtual machine will be created. This may be relative or absolute.
  If relative, the path is relative to the working directory when `packer`
  is executed. This directory must not exist or be empty prior to running the builder.
  By default this is "output-BUILDNAME" where "BUILDNAME" is the name
  of the build.

* `qemuargs` (array of array of strings) - Allows complete control over
  the qemu command line (though not, at this time, qemu-img). Each array
  of strings makes up a command line switch that overrides matching default
  switch/value pairs. Any value specified as an empty string is ignored.
  All values after the switch are concatenated with no separater.

  WARNING: The qemu command line allows extreme flexibility, so beware of
  conflicting arguments causing failures of your run. For instance, using
   --no-acpi could break the ability to send power signal type commands (e.g.,
  shutdown -P now) to the virtual machine, thus preventing proper shutdown. To
  see the defaults, look in the packer.log file and search for the
  qemu-system-x86 command. The arguments are all printed for review.

  The following shows a sample usage:

<pre class="prettyprint">
  . . .
  "qemuargs": [
    [ "-m", "1024m" ],
    [ "--no-acpi", "" ],
    [
       "-netdev",
      "user,id=mynet0,",
      "hostfwd=hostip:hostport-guestip:guestport",
      ""
    ],
    [ "-device", "virtio-net,netdev=mynet0" ]
  ]
  . . .
</pre>

  would produce the following (not including other defaults supplied by the builder and not otherwise conflicting with the qemuargs):

<pre class="prettyprint">
    qemu-system-x86 -m 1024m --no-acpi -netdev user,id=mynet0,hostfwd=hostip:hostport-guestip:guestport -device virtio-net,netdev=mynet0"
</pre>

* `qemu_binary` (string) - The name of the Qemu binary to look for.  This
  defaults to "qemu-system-x86_64", but may need to be changed for some
  platforms.  For example "qemu-kvm", or "qemu-system-i386" may be a better
  choice for some systems.

* `shutdown_command` (string) - The command to use to gracefully shut down
  the machine once all the provisioning is done. By default this is an empty
  string, which tells Packer to just forcefully shut down the machine.

* `shutdown_timeout` (string) - The amount of time to wait after executing
  the `shutdown_command` for the virtual machine to actually shut down.
  If it doesn't shut down in this time, it is an error. By default, the timeout
  is "5m", or five minutes.

* `ssh_host_port_min` and `ssh_host_port_max` (uint) - The minimum and
  maximum port to use for the SSH port on the host machine which is forwarded
  to the SSH port on the guest machine. Because Packer often runs in parallel,
  Packer will choose a randomly available port in this range to use as the
  host port.

* `ssh_key_path` (string) - Path to a private key to use for authenticating
  with SSH. By default this is not set (key-based auth won't be used).
  The associated public key is expected to already be configured on the
  VM being prepared by some other process (kickstart, etc.).

* `ssh_password` (string) - The password for `ssh_username` to use to
  authenticate with SSH. By default this is the empty string.

* `ssh_port` (integer) - The port that SSH will be listening on in the guest
  virtual machine. By default this is 22. The Qemu builder will map, via
  port forward, a port on the host machine to the port listed here so
  machines outside the installing VM can access the VM.

* `ssh_wait_timeout` (string) - The duration to wait for SSH to become
  available. By default this is "20m", or 20 minutes. Note that this should
  be quite long since the timer begins as soon as the virtual machine is booted.

* `vm_name` (string) - This is the name of the image (QCOW2 or IMG) file for
  the new virtual machine, without the file extension. By default this is
  "packer-BUILDNAME", where "BUILDNAME" is the name of the build.

* `vnc_port_min` and `vnc_port_max` (integer) - The minimum and
  maximum port to use for the VNC port on the host machine which is forwarded
  to the VNC port on the guest machine. Because Packer often runs in parallel,
  Packer will choose a randomly available port in this range to use as the
  host port.

## Boot Command

The `boot_command` configuration is very important: it specifies the keys
to type when the virtual machine is first booted in order to start the
OS installer. This command is typed after `boot_wait`, which gives the
virtual machine some time to actually load the ISO.

As documented above, the `boot_command` is an array of strings. The
strings are all typed in sequence. It is an array only to improve readability
within the template.

The boot command is "typed" character for character over a VNC connection
to the machine, simulating a human actually typing the keyboard. There are
a set of special keys available. If these are in your boot command, they
will be replaced by the proper key:

* `<enter>` and `<return>` - Simulates an actual "enter" or "return" keypress.

* `<esc>` - Simulates pressing the escape key.

* `<tab>` - Simulates pressing the tab key.

* `<wait>` `<wait5>` `<wait10>` - Adds a 1, 5 or 10 second pause before sending any additional keys. This
  is useful if you have to generally wait for the UI to update before typing more.

In addition to the special keys, each command to type is treated as a
[configuration template](/docs/templates/configuration-templates.html).
The available variables are:

* `HTTPIP` and `HTTPPort` - The IP and port, respectively of an HTTP server
  that is started serving the directory specified by the `http_directory`
  configuration parameter. If `http_directory` isn't specified, these will
  be blank!

Example boot command. This is actually a working boot command used to start
an CentOS 6.4 installer:

<pre class="prettyprint">
"boot_command":
[
  "<tab><wait>",
  " ks=http://10.0.2.2:{{ .HTTPPort }}/centos6-ks.cfg<enter>"
]
</pre>
