---
layout: "docs"
---

# VirtualBox Builder

Type: `virtualbox`

The VirtualBox builder is able to create [VirtualBox](https://www.virtualbox.org/)
virtual machines and export them in the OVF format.

The builder builds a virtual machine by creating a new virtual machine
from scratch, booting it, installing an OS, provisioning software within
the OS, then shutting it down. The result of the VirtualBox builder is a directory
containing all the files necessary to run the virtual machine portably.

## Basic Example

Here is a basic example. This example is not functional. It will start the
OS installer but then fail because we don't provide the preseed file for
Ubuntu to self-install. Still, the example serves to show the basic configuration:

<pre class="prettyprint">
{
  "type": "virtualbox",
  "guest_os_type": "Ubuntu_64",
  "iso_url": "http://releases.ubuntu.com/12.04/ubuntu-12.04.2-server-amd64.iso",
  "iso_checksum": "af5f788aee1b32c4b2634734309cc9e9",
  "iso_checksum_type": "md5",
  "ssh_username": "packer",
  "ssh_wait_timeout": "30s",
  "shutdown_command": "shutdown -P now"
}
</pre>

## Configuration Reference

There are many configuration options available for the VirtualBox builder.
They are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

Required:

* `iso_checksum` (string) - The checksum for the OS ISO file. Because ISO
  files are so large, this is required and Packer will verify it prior
  to booting a virtual machine with the ISO attached. The type of the
  checksum is specified with `iso_checksum_type`, documented below.

* `iso_checksum_type` (string) - The type of the checksum specified in
  `iso_checksum`. Valid values are "md5", "sha1", or "sha256" currently.

* `iso_url` (string) - A URL to the ISO containing the installation image.
  This URL can be either an HTTP URL or a file URL (or path to a file).
  If this is an HTTP URL, Packer will download it and cache it between
  runs.

* `ssh_username` (string) - The username to use to SSH into the machine
  once the OS is installed.

Optional:

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

* `disk_size` (int) - The size, in megabytes, of the hard disk to create
  for the VM. By default, this is 40000 (40 GB).

* `floppy_files` (array of strings) - A list of files to put onto a floppy
  disk that is attached when the VM is booted for the first time. This is
  most useful for unattended Windows installs, which look for an
  `Autounattend.xml` file on removable media. By default no floppy will
  be attached. The files listed in this configuration will all be put
  into the root directory of the floppy disk; sub-directories are not supported.

* `guest_additions_path` (string) - The path on the guest virtual machine
  where the VirtualBox guest additions ISO will be uploaded. By default this
  is "VBoxGuestAdditions.iso" which should upload into the login directory
  of the user. This is a [configuration template](/docs/templates/configuration-templates.html)
  where the `Version` variable is replaced with the VirtualBox version.

* `guest_additions_sha256` (string) - The SHA256 checksum of the guest
  additions ISO that will be uploaded to the guest VM. By default the
  checksums will be downloaded from the VirtualBox website, so this only
  needs to be set if you want to be explicit about the checksum.

* `guest_additions_url` (string) - The URL to the guest additions ISO
  to upload. This can also be a file URL if the ISO is at a local path.
  By default the VirtualBox builder will go and download the proper
  guest additions ISO from the internet.

* `guest_os_type` (string) - The guest OS type being installed. By default
  this is "other", but you can get _dramatic_ performance improvements by
  setting this to the proper value. To view all available values for this
  run `VBoxManage list ostypes`. Setting the correct value hints to VirtualBox
  how to optimize the virtual hardware to work best with that operating
  system.

* `headless` (bool) - Packer defaults to building VirtualBox
  virtual machines by launching a GUI that shows the console of the
  machine being built. When this value is set to true, the machine will
  start without a console.

* `http_directory` (string) - Path to a directory to serve using an HTTP
  server. The files in this directory will be available over HTTP that will
  be requestable from the virtual machine. This is useful for hosting
  kickstart files and so on. By default this is "", which means no HTTP
  server will be started. The address and port of the HTTP server will be
  available as variables in `boot_command`. This is covered in more detail
  below.

* `http_port_min` and `http_port_max` (int) - These are the minimum and
  maximum port to use for the HTTP server started to serve the `http_directory`.
  Because Packer often runs in parallel, Packer will choose a randomly available
  port in this range to run the HTTP server. If you want to force the HTTP
  server to be on one port, make this minimum and maximum port the same.
  By default the values are 8000 and 9000, respectively.

* `output_directory` (string) - This is the path to the directory where the
  resulting virtual machine will be created. This may be relative or absolute.
  If relative, the path is relative to the working directory when `packer`
  is executed. This directory must not exist or be empty prior to running the builder.
  By default this is "output-BUILDNAME" where "BUILDNAME" is the name
  of the build.

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

* `ssh_password` (string) - The password for `ssh_username` to use to
  authenticate with SSH. By default this is the empty string.

* `ssh_port` (int) - The port that SSH will be listening on in the guest
  virtual machine. By default this is 22.

* `ssh_wait_timeout` (string) - The duration to wait for SSH to become
  available. By default this is "20m", or 20 minutes. Note that this should
  be quite long since the timer begins as soon as the virtual machine is booted.

* `vboxmanage` (array of array of strings) - Custom `VBoxManage` commands to
  execute in order to further customize the virtual machine being created.
  The value of this is an array of commands to execute. The commands are executed
  in the order defined in the template. For each command, the command is
  defined itself as an array of strings, where each string represents a single
  argument on the command-line to `VBoxManage` (but excluding `VBoxManage`
  itself). Each arg is treated as a [configuration template](/docs/templates/configuration-templates.html),
  where the `Name` variable is replaced with the VM name. More details on how
  to use `VBoxManage` are below.

* `virtualbox_version_file` (string) - The path within the virtual machine
  to upload a file that contains the VirtualBox version that was used to
  create the machine. This information can be useful for provisioning.
  By default this is ".vbox_version", which will generally upload it into
  the home directory.

* `vm_name` (string) - This is the name of the VMX file for the new virtual
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
an Ubuntu 12.04 installer:

<pre class="prettyprint">
[
  "&lt;esc&gt;&lt;esc&gt;&lt;enter&gt;&lt;wait&gt;",
  "/install/vmlinuz noapic ",
  "preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg ",
  "debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
  "hostname={{ .Name }} ",
  "fb=false debconf/frontend=noninteractive ",
  "keyboard-configuration/modelcode=SKIP keyboard-configuration/layout=USA ",
  "keyboard-configuration/variant=USA console-setup/ask_detect=false ",
  "initrd=/install/initrd.gz -- &lt;enter&gt;"
]
</pre>

## Guest Additions

Packer will automatically download the proper guest additions for the
version of VirtualBox that is running and upload those guest additions into
the virtual machine so that provisioners can easily install them.

Packer downloads the guest additions from the official VirtualBox website,
and verifies the file with the official checksums released by VirtualBox.

After the virtual machine is up and the operating system is installed,
Packer uploads the guest additions into the virtual machine. The path where
they are uploaded is controllable by `guest_additions_path`, and defaults
to "VBoxGuestAdditions.iso". Without an absolute path, it is uploaded to the
home directory of the SSH user.

## VBoxManage Commands

In order to perform extra customization of the virtual machine, a template
can define extra calls to `VBoxMangage` to perform. [VBoxManage](http://www.virtualbox.org/manual/ch08.html)
is the command-line interface to VirtualBox where you can completely control
VirtualBox. It can be used to do things such as set RAM, CPUs, etc.

Extra VBoxManage commands are defined in the template in the `vboxmanage` section.
An example is shown below that sets the memory and number of CPUs within the
virtual machine:

<pre class="prettyprint">
{
  "vboxmanage": [
    ["modifyvm", "{{.Name}}", "--memory", "1024"],
    ["modifyvm", "{{.Name}}", "--cpus", "2"]
  ]
}
</pre>

The value of `vboxmanage` is an array of commands to execute. These commands
are executed in the order defined. So in the above example, the memory will be
set followed by the CPUs.

Each command itself is an array of strings, where each string is an argument
to `VBoxManage`. Each argument is treated as a
[configuration template](/docs/templates/configuration-templates.html).
The only available variable is `Name` which is replaced with the unique
name of the VM, which is required for many VBoxManage calls.
