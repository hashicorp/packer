---
layout: "docs"
page_title: "Amazon AMI Builder (chroot)"
---

# AMI Builder (chroot)

Type: `amazon-chroot`

The `amazon-chroot` builder is able to create Amazon AMIs backed by
an EBS volume as the root device. For more information on the difference
between instance storage and EBS-backed instances, see the
["storage for the root device" section in the EC2 documentation](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ComponentsAMIs.html#storage-for-the-root-device).

The difference between this builder and the `amazon-ebs` builder is that
this builder is able to build an EBS-backed AMI without launching a new
EC2 instance. This can dramatically speed up AMI builds for organizations
who need the extra fast build.

<div class="alert alert-block alert-warn">
<p><strong>This is an advanced builder.</strong> If you're just getting
started with Packer, we recommend starting with the
<a href="/docs/builders/amazon-ebs.html">amazon-ebs builder</a>, which is
much easier to use.</p>
</div>

The builder does _not_ manage AMIs. Once it creates an AMI and stores it
in your account, it is up to you to use, delete, etc. the AMI.

## How Does it Work?

This builder works by creating a new EBS volume from an existing source AMI
and attaching it into an already-running EC2 instance. One attached, a
[chroot](http://en.wikipedia.org/wiki/Chroot) is used to provision the
system within that volume. After provisioning, the volume is detached,
snapshotted, and an AMI is made.

Using this process, minutes can be shaved off the AMI creation process
because a new EC2 instance doesn't need to be launched.

There are some restrictions, however. The host EC2 instance where the
volume is attached to must be a similar system (generally the same OS
version, kernel versions, etc.) as the AMI being built. Additionally,
this process is much more expensive because the EC2 instance must be kept
running persistently in order to build AMIs, whereas the other AMI builders
start instances on-demand to build AMIs as needed.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

### Required:

* `access_key` (string) - The access key used to communicate with AWS.
  If not specified, Packer will use the environment variables
  `AWS_ACCESS_KEY_ID` or `AWS_ACCESS_KEY` (in that order), if set.
  If the environmental variables aren't set and Packer is running on
  an EC2 instance, Packer will check the instance metadata for IAM role
  keys.

* `ami_name` (string) - The name of the resulting AMI that will appear
  when managing AMIs in the AWS console or via APIs. This must be unique.
  To help make this unique, use a function like `timestamp` (see
  [configuration templates](/docs/templates/configuration-templates.html) for more info)

* `secret_key` (string) - The secret key used to communicate with AWS.
  If not specified, Packer will use the environment variables
  `AWS_SECRET_ACCESS_KEY` or `AWS_SECRET_KEY` (in that order), if set.
  If the environmental variables aren't set and Packer is running on
  an EC2 instance, Packer will check the instance metadata for IAM role
  keys.

* `source_ami` (string) - The source AMI whose root volume will be copied
  and provisioned on the currently running instance. This must be an
  EBS-backed AMI with a root volume snapshot that you have access to.

### Optional:

* `ami_description` (string) - The description to set for the resulting
  AMI(s). By default this description is empty.

* `ami_groups` (array of strings) - A list of groups that have access
  to launch the resulting AMI(s). By default no groups have permission
  to launch the AMI. `all` will make the AMI publicly accessible.

* `ami_product_codes` (array of strings) - A list of product codes to
  associate with the AMI. By default no product codes are associated with
  the AMI.

* `ami_regions` (array of strings) - A list of regions to copy the AMI to.
  Tags and attributes are copied along with the AMI. AMI copying takes time
  depending on the size of the AMI, but will generally take many minutes.

* `ami_users` (array of strings) - A list of account IDs that have access
  to launch the resulting AMI(s). By default no additional users other than the user
  creating the AMI has permissions to launch it.

* `ami_virtualization_type` (string) - The type of virtualization for the AMI
  you are building. This option is required to register HVM images. Can be
  "paravirtual" (default) or "hvm".

* `chroot_mounts` (array of array of strings) - This is a list of additional
  devices to mount into the chroot environment. This configuration parameter
  requires some additional documentation which is in the "Chroot Mounts" section
  below. Please read that section for more information on how to use this.

* `command_wrapper` (string) - How to run shell commands. This
  defaults to "{{.Command}}". This may be useful to set if you want to set
  environmental variables or perhaps run it with `sudo` or so on. This is a
  configuration template where the `.Command` variable is replaced with the
  command to be run.

* `copy_files` (array of strings) - Paths to files on the running EC2 instance
  that will be copied into the chroot environment prior to provisioning.
  This is useful, for example, to copy `/etc/resolv.conf` so that DNS lookups
  work.

* `device_path` (string) - The path to the device where the root volume
  of the source AMI will be attached. This defaults to "" (empty string),
  which forces Packer to find an open device automatically.

* `mount_path` (string) - The path where the volume will be mounted. This is
  where the chroot environment will be. This defaults to
  `packer-amazon-chroot-volumes/{{.Device}}`. This is a configuration
  template where the `.Device` variable is replaced with the name of the
  device where the volume is attached.

* `tags` (object of key/value strings) - Tags applied to the AMI.

## Basic Example

Here is a basic example. It is completely valid except for the access keys:

<pre class="prettyprint">
{
  "type": "amazon-chroot",
  "access_key": "YOUR KEY HERE",
  "secret_key": "YOUR SECRET KEY HERE",
  "source_ami": "ami-e81d5881",
  "ami_name": "packer-amazon-chroot {{timestamp}}"
}
</pre>

## Chroot Mounts

The `chroot_mounts` configuration can be used to mount additional devices
within the chroot. By default, the following additional mounts are added
into the chroot by Packer:

* `/proc` (proc)
* `/sys` (sysfs)
* `/dev` (bind to real `/dev`)
* `/dev/pts` (devpts)
* `/proc/sys/fs/binfmt_misc` (binfmt_misc)

These default mounts are usually good enough for anyone and are sane
defaults. However, if you want to change or add the mount points, you may
using the `chroot_mounts` configuration. Here is an example configuration:

<pre class="prettyprint">
{
  "chroot_mounts": [
    ["proc", "proc", "/proc"],
    ["bind", "/dev", "/dev"]
  ]
}
</pre>

`chroot_mounts` is a list of a 3-tuples of strings. The three components
of the 3-tuple, in order, are:

* The filesystem type. If this is "bind", then Packer will properly bind
  the filesystem to another mount point.

* The source device.

* The mount directory.

## Parallelism

A quick note on parallelism: it is perfectly safe to run multiple
_separate_ Packer processes with the `amazon-chroot` builder on the same
EC2 instance. In fact, this is recommended as a way to push the most performance
out of your AMI builds.

Packer properly obtains a process lock for the parallelism-sensitive parts
of its internals such as finding an available device.

## Gotchas

One of the difficulties with using the chroot builder is that your provisioning
scripts must not leave any processes running or packer will be unable to unmount
the filesystem.

For debian based distributions you can setup a [policy-rc.d](http://people.debian.org/~hmh/invokerc.d-policyrc.d-specification.txt) file which will
prevent packages installed by your provisioners from starting services:

<pre class="prettyprint">
{
  "type": "shell",
  "inline": [
    "echo '#!/bin/sh' > /usr/sbin/policy-rc.d",
    "echo 'exit 101' >> /usr/sbin/policy-rc.d",
    "chmod a+x /usr/sbin/policy-rc.d"
  ]
},
# ...
{
  "type": "shell",
  "inline": [
    "rm -f /usr/sbin/policy-rc.d"
  ]
}
</pre>
