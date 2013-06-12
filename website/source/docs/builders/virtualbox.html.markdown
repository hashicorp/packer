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
  "iso_md5": "af5f788aee1b32c4b2634734309cc9e9",
  "ssh_username": "packer",
  "ssh_wait_timeout": "30s"
}
</pre>

## Configuration Reference

There are many configuration options available for the VirtualBox builder.
They are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

Required:

* `iso_md5` (string) - The MD5 checksum for the OS ISO file. Because ISO
  files are so large, this is required and Packer will verify it prior
  to booting a virtual machine with the ISO attached.

* `iso_url` (string) - A URL to the ISO containing the installation image.
  This URL can be either an HTTP URL or a file URL (or path to a file).
  If this is an HTTP URL, Packer will download it and cache it between
  runs.

* `ssh_username` (string) - The username to use to SSH into the machine
  once the OS is installed.

Optional:

* `boot_command` (array of strings) - This is an array of commands to type
  when the virtual machine is firsted booted. The goal of these commands should
  be to type just enough to initialize the operating system installer. Special
  keys can be typed as well, and are covered in the section below on the boot
  command. If this is not specified, it is assumed the installer will start
  itself.

* `boot_wait` (string) - The time to wait after booting the initial virtual
  machine before typing the `boot_command`. The value of this should be
  a duration. Examples are "5s" and "1m30s" which will cause Packer to wait
  five seconds and one minute 30 seconds, respectively. If this isn't specified,
  the default is 10 seconds.

* `guest_os_type` (string) - The guest OS type being installed. By default
  this is "other", but you can get _dramatic_ performance improvements by
  setting this to the proper value. To view all available values for this
  run `VBoxManage list ostypes`. Setting the correct value hints to VirtualBox
  how to optimize the virtual hardware to work best with that operating
  system.

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
  is executed. By default this is "virtualbox". This directory must not exist
  or be empty prior to running the builder.

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

* `ssh_wait_timeout` (string) - The duration to wait for SSH to become
  available. By default this is "20m", or 20 minutes. Note that this should
  be quite long since the timer begins as soon as virtual machine is booted.

* `vm_name` (string) - This is the name of the VMX file for the new virtual
  machine, without the file extension. By default this is "packer".

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

* `<wait>` - Adds a one second pause before sending any additional keys. This
  is useful if you have to generally wait for the UI to update before typing more.

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
  "initrd=/install/initrd.gz -- <enter>"
]
</pre>
