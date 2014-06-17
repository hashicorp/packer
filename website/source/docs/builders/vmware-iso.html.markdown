---
layout: "docs"
page_title: "VMware Builder from ISO"
---

# VMware Builder (from ISO)

Type: `vmware-iso`

This VMware builder is able to create VMware virtual machines from an
ISO file as a source. It currently
supports building virtual machines on hosts running
[VMware Fusion](http://www.vmware.com/products/fusion/overview.html) for OS X,
[VMware Workstation](http://www.vmware.com/products/workstation/overview.html)
for Linux and Windows, and
[VMware Player](http://www.vmware.com/products/player/) on Linux. It can
also build machines directly on
[VMware vSphere Hypervisor](http://www.vmware.com/products/vsphere-hypervisor/)
using SSH as opposed to the vSphere API.

The builder builds a virtual machine by creating a new virtual machine
from scratch, booting it, installing an OS, provisioning software within
the OS, then shutting it down. The result of the VMware builder is a directory
containing all the files necessary to run the virtual machine.

## Basic Example

Here is a basic example. This example is not functional. It will start the
OS installer but then fail because we don't provide the preseed file for
Ubuntu to self-install. Still, the example serves to show the basic configuration:

<pre class="prettyprint">
{
  "type": "vmware-iso",
  "iso_url": "http://old-releases.ubuntu.com/releases/precise/ubuntu-12.04.2-server-amd64.iso",
  "iso_checksum": "af5f788aee1b32c4b2634734309cc9e9",
  "iso_checksum_type": "md5",
  "ssh_username": "packer",
  "ssh_wait_timeout": "30s",
  "shutdown_command": "shutdown -P now"
}
</pre>

## Configuration Reference

There are many configuration options available for the VMware builder.
They are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

### Required:

* `iso_checksum` (string) - The checksum for the OS ISO file. Because ISO
  files are so large, this is required and Packer will verify it prior
  to booting a virtual machine with the ISO attached. The type of the
  checksum is specified with `iso_checksum_type`, documented below.

* `iso_checksum_type` (string) - The type of the checksum specified in
  `iso_checksum`. Valid values are "none", "md5", "sha1", "sha256", or
  "sha512" currently. While "none" will skip checksumming, this is not
  recommended since ISO files are generally large and corruption does happen
  from time to time.

* `iso_url` (string) - A URL to the ISO containing the installation image.
  This URL can be either an HTTP URL or a file URL (or path to a file).
  If this is an HTTP URL, Packer will download it and cache it between
  runs.

* `ssh_username` (string) - The username to use to SSH into the machine
  once the OS is installed.

### Optional:

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

* `disk_size` (integer) - The size of the hard disk for the VM in megabytes.
  The builder uses expandable, not fixed-size virtual hard disks, so the
  actual file representing the disk will not use the full size unless it is full.
  By default this is set to 40,000 (about 40 GB).

* `disk_type_id` (string) - The type of VMware virtual disk to create.
  The default is "1", which corresponds to a growable virtual disk split in
  2GB files.  This option is for advanced usage, modify only if you
  know what you're doing.  For more information, please consult the
  [Virtual Disk Manager User's Guide](http://www.vmware.com/pdf/VirtualDiskManager.pdf)
  for desktop VMware clients. For ESXi, refer to the proper ESXi documentation.

* `floppy_files` (array of strings) - A list of files to place onto a floppy
  disk that is attached when the VM is booted. This is most useful
  for unattended Windows installs, which look for an `Autounattend.xml` file
  on removable media. By default, no floppy will be attached. All files
  listed in this setting get placed into the root directory of the floppy
  and the floppy is attached as the first floppy device. Currently, no
  support exists for creating sub-directories on the floppy. Wildcard
  characters (*, ?, and []) are allowed. Directory names are also allowed,
  which will add all the files found in the directory to the floppy.

* `fusion_app_path` (string) - Path to "VMware Fusion.app". By default this
  is "/Applications/VMware Fusion.app" but this setting allows you to
  customize this.

* `guest_os_type` (string) - The guest OS type being installed. This will be
  set in the VMware VMX. By default this is "other". By specifying a more specific
  OS type, VMware may perform some optimizations or virtual hardware changes
  to better support the operating system running in the virtual machine.

* `headless` (boolean) - Packer defaults to building VMware
  virtual machines by launching a GUI that shows the console of the
  machine being built. When this value is set to true, the machine will
  start without a console. For VMware machines, Packer will output VNC
  connection information in case you need to connect to the console to
  debug the build process.

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

* `output_directory` (string) - This is the path to the directory where the
  resulting virtual machine will be created. This may be relative or absolute.
  If relative, the path is relative to the working directory when `packer`
  is executed. This directory must not exist or be empty prior to running the builder.
  By default this is "output-BUILDNAME" where "BUILDNAME" is the name
  of the build.

* `remote_datastore` (string) - The path to the datastore where the resulting
  VM will be stored when it is built on the remote machine. By default this
  is "datastore1". This only has an effect if `remote_type` is enabled.

* `remote_host` (string) - The host of the remote machine used for access.
  This is only required if `remote_type` is enabled.

* `remote_password` (string) - The SSH password for the user used to
  access the remote machine. By default this is empty. This only has an
  effect if `remote_type` is enabled.

* `remote_type` (string) - The type of remote machine that will be used to
  build this VM rather than a local desktop product. The only value accepted
  for this currently is "esx5". If this is not set, a desktop product will be
  used. By default, this is not set.

* `remote_username` (string) - The username for the SSH user that will access
  the remote machine. This is required if `remote_type` is enabled.

* `shutdown_command` (string) - The command to use to gracefully shut down
  the machine once all the provisioning is done. By default this is an empty
  string, which tells Packer to just forcefully shut down the machine.

* `shutdown_timeout` (string) - The amount of time to wait after executing
  the `shutdown_command` for the virtual machine to actually shut down.
  If it doesn't shut down in this time, it is an error. By default, the timeout
  is "5m", or five minutes.

* `skip_compaction` (boolean) -  VMware-created disks are defragmented
  and compacted at the end of the build process using `vmware-vdiskmanager`.
  In certain rare cases, this might actually end up making the resulting disks
  slightly larger. If you find this to be the case, you can disable compaction
  using this configuration value.

* `ssh_host` (string) - Hostname or IP address of the host. By default, DHCP
  is used to connect to the host and this field is not used.

* `ssh_key_path` (string) - Path to a private key to use for authenticating
  with SSH. By default this is not set (key-based auth won't be used).
  The associated public key is expected to already be configured on the
  VM being prepared by some other process (kickstart, etc.).

* `ssh_password` (string) - The password for `ssh_username` to use to
  authenticate with SSH. By default this is the empty string.

* `ssh_port` (integer) - The port that SSH will listen on within the virtual
  machine. By default this is 22.

* `ssh_skip_request_pty` (boolean) - If true, a pty will not be requested as
  part of the SSH connection. By default, this is "false", so a pty
  _will_ be requested.

* `ssh_wait_timeout` (string) - The duration to wait for SSH to become
  available. By default this is "20m", or 20 minutes. Note that this should
  be quite long since the timer begins as soon as the virtual machine is booted.

* `tools_upload_flavor` (string) - The flavor of the VMware Tools ISO to
  upload into the VM. Valid values are "darwin", "linux", and "windows".
  By default, this is empty, which means VMware tools won't be uploaded.

* `tools_upload_path` (string) - The path in the VM to upload the VMware
  tools. This only takes effect if `tools_upload_flavor` is non-empty.
  This is a [configuration template](/docs/templates/configuration-templates.html)
  that has a single valid variable: `Flavor`, which will be the value of
  `tools_upload_flavor`. By default the upload path is set to
  `{{.Flavor}}.iso`.

* `vm_name` (string) - This is the name of the VMX file for the new virtual
  machine, without the file extension. By default this is "packer-BUILDNAME",
  where "BUILDNAME" is the name of the build.

* `vmdk_name` (string) - The filename of the virtual disk that'll be created,
  without the extension. This defaults to "packer".

* `vmx_data` (object of key/value strings) - Arbitrary key/values
  to enter into the virtual machine VMX file. This is for advanced users
  who want to set properties such as memory, CPU, etc.

* `vmx_data_post` (object of key/value strings) - Identical to `vmx_data`,
  except that it is run after the virtual machine is shutdown, and before the
  virtual machine is exported.

* `vmx_template_path` (string) - Path to a
  [configuration template](/docs/templates/configuration-templates.html) that
  defines the contents of the virtual machine VMX file for VMware. This is
  for **advanced users only** as this can render the virtual machine
  non-functional. See below for more information. For basic VMX modifications,
  try `vmx_data` first.

* `vnc_port_min` and `vnc_port_max` (integer) - The minimum and maximum port to
  use for VNC access to the virtual machine. The builder uses VNC to type
  the initial `boot_command`. Because Packer generally runs in parallel, Packer
  uses a randomly chosen port in this range that appears available. By default
  this is 5900 to 6000. The minimum and maximum ports are inclusive.

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

## VMX Template

The heart of a VMware machine is the "vmx" file. This contains all the
virtual hardware metadata necessary for the VM to function. Packer by default
uses a [safe, flexible VMX file](https://github.com/mitchellh/packer/blob/20541a7eda085aa5cf35bfed5069592ca49d106e/builder/vmware/step_create_vmx.go#L84).
But for advanced users, this template can be customized. This allows
Packer to build virtual machines of effectively any guest operating system
type.

<div class="alert alert-block alert-warn">
<p>
<strong>This is an advanced feature.</strong> Modifying the VMX template
can easily cause your virtual machine to not boot properly. Please only
modify the template if you know what you're doing.
</p>
</div>

Within the template, a handful of variables are available so that your
template can continue working with the rest of the Packer machinery. Using
these variables isn't required, however.

* `Name` - The name of the virtual machine.
* `GuestOS` - The VMware-valid guest OS type.
* `DiskName` - The filename (without the suffix) of the main virtual disk.
* `ISOPath` - The path to the ISO to use for the OS installation.

## Building on a Remote vSphere Hypervisor

In addition to using the desktop products of VMware locally to build
virtual machines, Packer can use a remote VMware Hypervisor to build
the virtual machine.

<div class="alert alert-block alert-info">
Note: Packer supports ESXi 5.1 and above.
</div>

Before using a remote vSphere Hypervisor, you need to enable GuestIPHack by running the following command:

```
esxcli system settings advanced set -o /Net/GuestIPHack -i 1
```

When using a remote VMware Hypervisor, the builder still downloads the
ISO and various files locally, and uploads these to the remote machine.
Packer currently uses SSH to communicate to the ESXi machine rather than
the vSphere API. At some point, the vSphere API may be used.

To use a remote VMware vSphere Hypervisor to build your virtual machine,
fill in the required `remote_*` configurations:

* `remote_type` - This must be set to "esx5".

* `remote_host` - The host of the remote machine.

Additionally, there are some optional configurations that you'll likely
have to modify as well:

* `remote_datastore` - The path to the datastore where the VM will be
  stored on the ESXi machine.

* `remote_username` - The SSH username used to access the remote machine.

* `remote_password` - The SSH password for access to the remote machine.
