---
layout: "docs"
page_title: "VMware Builder from VMX"
---

# VMware Builder (from VMX)

Type: `vmware-vmx`

This VMware builder is able to create VMware virtual machines from an
existing VMware virtual machine (a VMX file). It currently
supports building virtual machines on hosts running
[VMware Fusion](http://www.vmware.com/products/fusion/overview.html) for OS X,
[VMware Workstation](http://www.vmware.com/products/workstation/overview.html)
for Linux and Windows, and
[VMware Player](http://www.vmware.com/products/player/) on Linux.

The builder builds a virtual machine by cloning the VMX file using
the clone capabilities introduced in VMware Fusion 6, Workstation 10,
and Player 6. After cloning the VM, it provisions software within the
new machine, shuts it down, and compacts the disks. The resulting folder
contains a new VMware virtual machine.

## Basic Example

Here is an example. This example is fully functional as long as the source
path points to a real VMX file with the proper settings:

<pre class="prettyprint">
{
  "type": "vmware-vmx",
  "source_path": "/path/to/a/vm.vmx",
  "ssh_username": "root",
  "ssh_password": "root",
  "shutdown_command": "shutdown -P now"
}
</pre>

## Configuration Reference

There are many configuration options available for the VMware builder.
They are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

Required:

* `source_path` (string) - Path to the source VMX file to clone.

* `ssh_username` (string) - The username to use to SSH into the machine
  once the OS is installed.

Optional:

* `boot_wait` (string) - The time to wait after booting the initial virtual
  machine before typing the `boot_command`. The value of this should be
  a duration. Examples are "5s" and "1m30s" which will cause Packer to wait
  five seconds and one minute 30 seconds, respectively. If this isn't specified,
  the default is 10 seconds.

* `floppy_files` (array of strings) - A list of files to put onto a floppy
  disk that is attached when the VM is booted for the first time. This is
  most useful for unattended Windows installs, which look for an
  `Autounattend.xml` file on removable media. By default no floppy will
  be attached. The files listed in this configuration will all be put
  into the root directory of the floppy disk; sub-directories are not supported.

* `fusion_app_path` (string) - Path to "VMware Fusion.app". By default this
  is "/Applications/VMware Fusion.app" but this setting allows you to
  customize this.

* `headless` (bool) - Packer defaults to building VMware
  virtual machines by launching a GUI that shows the console of the
  machine being built. When this value is set to true, the machine will
  start without a console. For VMware machines, Packer will output VNC
  connection information in case you need to connect to the console to
  debug the build process.

* `output_directory` (string) - This is the path to the directory where the
  resulting virtual machine will be created. This may be relative or absolute.
  If relative, the path is relative to the working directory when `packer`
  is executed. This directory must not exist or be empty prior to running the builder.
  By default this is "output-BUILDNAME" where "BUILDNAME" is the name
  of the build.

* `skip_compaction` (bool) -  VMware-created disks are defragmented
  and compacted at the end of the build process using `vmware-vdiskmanager`.
  In certain rare cases, this might actually end up making the resulting disks
  slightly larger. If you find this to be the case, you can disable compaction
  using this configuration value.

* `shutdown_command` (string) - The command to use to gracefully shut down
  the machine once all the provisioning is done. By default this is an empty
  string, which tells Packer to just forcefully shut down the machine.

* `shutdown_timeout` (string) - The amount of time to wait after executing
  the `shutdown_command` for the virtual machine to actually shut down.
  If it doesn't shut down in this time, it is an error. By default, the timeout
  is "5m", or five minutes.

* `ssh_key_path` (string) - Path to a private key to use for authenticating
  with SSH. By default this is not set (key-based auth won't be used).
  The associated public key is expected to already be configured on the
  VM being prepared by some other process (kickstart, etc.).

* `ssh_password` (string) - The password for `ssh_username` to use to
  authenticate with SSH. By default this is the empty string.

* `ssh_port` (int) - The port that SSH will listen on within the virtual
  machine. By default this is 22.

* `ssh_skip_request_pty` (bool) - If true, a pty will not be requested as
  part of the SSH connection. By default, this is "false", so a pty
  _will_ be requested.

* `ssh_wait_timeout` (string) - The duration to wait for SSH to become
  available. By default this is "20m", or 20 minutes. Note that this should
  be quite long since the timer begins as soon as the virtual machine is booted.

* `vm_name` (string) - This is the name of the VMX file for the new virtual
  machine, without the file extension. By default this is "packer-BUILDNAME",
  where "BUILDNAME" is the name of the build.

* `vmx_data` (object, string keys and string values) - Arbitrary key/values
  to enter into the virtual machine VMX file. This is for advanced users
  who want to set properties such as memory, CPU, etc.
