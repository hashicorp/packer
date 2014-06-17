---
layout: "docs"
page_title: "VirtualBox Builder (from an OVF/OVA)"
---

# VirtualBox Builder (from an OVF/OVA)

Type: `virtualbox-ovf`

This VirtualBox builder is able to create [VirtualBox](https://www.virtualbox.org/)
virtual machines and export them in the OVF format, starting from an
existing OVF/OVA (exported virtual machine image).

The builder builds a virtual machine by importing an existing OVF or OVA
file. It then boots this image, runs provisioners on this new VM, and
exports that VM to create the image. The imported machine is deleted prior
to finishing the build.

## Basic Example

Here is a basic example. This example is functional if you have an OVF matching
the settings here.

<pre class="prettyprint">
{
  "type": "virtualbox-ovf",
  "source_path": "source.ovf",
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

There are many configuration options available for the VirtualBox builder.
They are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

### Required:

* `source_path` (string) - The path to an OVF or OVA file that acts as
  the source of this build.

* `ssh_username` (string) - The username to use to SSH into the machine
  once the OS is installed.

### Optional:

* `export_opts` (array of strings) - Additional options to pass to the `VBoxManage export`.
  This can be useful for passing product information to include in the resulting
  appliance file.

* `floppy_files` (array of strings) - A list of files to place onto a floppy
  disk that is attached when the VM is booted. This is most useful
  for unattended Windows installs, which look for an `Autounattend.xml` file
  on removable media. By default, no floppy will be attached. All files
  listed in this setting get placed into the root directory of the floppy
  and the floppy is attached as the first floppy device. Currently, no
  support exists for creating sub-directories on the floppy. Wildcard
  characters (*, ?, and []) are allowed. Directory names are also allowed,
  which will add all the files found in the directory to the floppy.

* `format` (string) - Either "ovf" or "ova", this specifies the output
  format of the exported virtual machine. This defaults to "ovf".

* `guest_additions_mode` (string) - The method by which guest additions
  are made available to the guest for installation. Valid options are
  "upload", "attach", or "disable". The functions of each of these should be
  self-explanatory. The default value is "upload". If "disable" is used,
  guest additions won't be downloaded, either.

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

* `headless` (boolean) - Packer defaults to building VirtualBox
  virtual machines by launching a GUI that shows the console of the
  machine being built. When this value is set to true, the machine will
  start without a console.

* `import_opts` (string) - Additional options to pass to the `VBoxManage import`.
  This can be useful for passing "keepallmacs" or "keepnatmacs" options for existing
  ovf images.

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

* `ssh_host_port_min` and `ssh_host_port_max` (integer) - The minimum and
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

* `vboxmanage_post` (array of array of strings) - Identical to `vboxmanage`,
  except that it is run after the virtual machine is shutdown, and before the
  virtual machine is exported.

* `virtualbox_version_file` (string) - The path within the virtual machine
  to upload a file that contains the VirtualBox version that was used to
  create the machine. This information can be useful for provisioning.
  By default this is ".vbox_version", which will generally be upload it into
  the home directory.

* `vm_name` (string) - This is the name of the virtual machine when it is
  imported as well as the name of the OVF file when the virtual machine is
  exported. By default this is "packer-BUILDNAME", where "BUILDNAME" is
  the name of the build.

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
can define extra calls to `VBoxManage` to perform. [VBoxManage](http://www.virtualbox.org/manual/ch08.html)
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
