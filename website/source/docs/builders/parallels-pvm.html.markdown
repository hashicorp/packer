---
layout: "docs"
page_title: "Parallels Builder (from a PVM)"
---

# Parallels Builder (from a PVM)

Type: `parallels-pvm`

This Parallels builder is able to create
[Parallels Desktop for Mac](http://www.parallels.com/products/desktop/)
virtual machines and export them in the PVM format, starting from an
existing PVM (exported virtual machine image).

The builder builds a virtual machine by importing an existing PVM
file. It then boots this image, runs provisioners on this new VM, and
exports that VM to create the image. The imported machine is deleted prior
to finishing the build.

## Basic Example

Here is a basic example. This example is functional if you have an PVM matching
the settings here.

<pre class="prettyprint">
{
  "type": "parallels-pvm",
  "source_path": "source.pvm",
  "ssh_username": "packer",
  "ssh_password": "packer",
  "ssh_wait_timeout": "30s",
  "shutdown_command": "echo 'packer' | sudo -S shutdown -P now"
}
</pre>

It is important to add a `shutdown_command`. By default Packer halts the
virtual machine and the file system may not be sync'd. Thus, changes made in a
provisioner might not be saved.

## Configuration Reference

There are many configuration options available for the Parallels builder.
They are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

### Required:

* `source_path` (string) - The path to a PVM directory that acts as
  the source of this build.

* `ssh_username` (string) - The username to use to SSH into the machine
  once the OS is installed.

### Optional:

* `floppy_files` (array of strings) - A list of files to put onto a floppy
  disk that is attached when the VM is booted for the first time. This is
  most useful for unattended Windows installs, which look for an
  `Autounattend.xml` file on removable media. By default no floppy will
  be attached. The files listed in this configuration will all be put
  into the root directory of the floppy disk; sub-directories are not supported.

* `output_directory` (string) - This is the path to the directory where the
  resulting virtual machine will be created. This may be relative or absolute.
  If relative, the path is relative to the working directory when `packer`
  is executed. This directory must not exist or be empty prior to running the builder.
  By default this is "output-BUILDNAME" where "BUILDNAME" is the name
  of the build.

* `parallels_tools_guest_path` (string) - The path on the guest virtual machine
  where the Parallels tools ISO will be uploaded. By default this is
  "prl-tools.iso" which should upload into the login directory of the user.
  This is a configuration template where the `Version` variable is replaced
  with the prlctl version.

* `parallels_tools_host_path` (string) - The path to the Parallels Tools ISO to
  upload. By default the Parallels builder will use the "other" OS tools ISO from
  the Parallels installation:
  "/Applications/Parallels Desktop.app/Contents/Resources/Tools/prl-tools-other.iso"

* `parallels_tools_mode` (string) - The method by which Parallels tools are
  made available to the guest for installation. Valid options are "upload",
  "attach", or "disable". The functions of each of these should be
  self-explanatory. The default value is "upload".

* `prlctl` (array of array of strings) - Custom `prlctl` commands to execute in
  order to further customize the virtual machine being created. The value of
  this is an array of commands to execute. The commands are executed in the order
  defined in the template. For each command, the command is defined itself as an
  array of strings, where each string represents a single argument on the
  command-line to `prlctl` (but excluding `prlctl` itself). Each arg is treated
  as a [configuration template](/docs/templates/configuration-templates.html),
  where the `Name` variable is replaced with the VM name. More details on how
  to use `prlctl` are below.

* `prlctl_version_file` (string) - The path within the virtual machine to upload
  a file that contains the `prlctl` version that was used to create the machine.
  This information can be useful for provisioning. By default this is
  ".prlctl_version", which will generally upload it into the home directory.

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

* `ssh_port` (integer) - The port that SSH will be listening on in the guest
  virtual machine. By default this is 22.

* `ssh_wait_timeout` (string) - The duration to wait for SSH to become
  available. By default this is "20m", or 20 minutes. Note that this should
  be quite long since the timer begins as soon as the virtual machine is booted.

* `vm_name` (string) - This is the name of the virtual machine when it is
  imported as well as the name of the PVM directory when the virtual machine is
  exported. By default this is "packer-BUILDNAME", where "BUILDNAME" is
  the name of the build.

## prlctl Commands
In order to perform extra customization of the virtual machine, a template can
define extra calls to `prlctl` to perform.
[prlctl](http://download.parallels.com/desktop/v4/wl/docs/en/Parallels_Command_Line_Reference_Guide/)
is the command-line interface to Parallels. It can be used to do things such as
set RAM, CPUs, etc.

Extra `prlctl` commands are defined in the template in the `prlctl` section.
An example is shown below that sets the memory and number of CPUs within the
virtual machine:

<pre class="prettyprint">
{
  "prlctl": [
    ["set", "{{.Name}}", "--memsize", "1024"],
    ["set", "{{.Name}}", "--cpus", "2"]
  ]
}
</pre>

The value of `prlctl` is an array of commands to execute. These commands are
executed in the order defined. So in the above example, the memory will be set
followed by the CPUs.

Each command itself is an array of strings, where each string is an argument to
`prlctl`. Each argument is treated as a
[configuration template](/docs/templates/configuration-templates.html). The only
available variable is `Name` which is replaced with the unique name of the VM,
which is required for many `prlctl` calls.
