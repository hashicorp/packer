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

Required:

* `access_key` (string) - The access key used to communicate with AWS.
  If not specified, Packer will attempt to read this from environmental
  variables `AWS_ACCESS_KEY_ID` or `AWS_ACCESS_KEY` (in that order).
  If the environmental variables aren't set and Packer is running on
  an EC2 instance, Packer will check the instance metadata for IAM role
  keys.

* `ami_name` (string) - The name of the resulting AMI that will appear
  when managing AMIs in the AWS console or via APIs. This must be unique.
  To help make this unique, use a function like `timestamp` (see
  [configuration templates](/docs/templates/configuration-templates.html) for more info)

* `secret_key` (string) - The secret key used to communicate with AWS.
  If not specified, Packer will attempt to read this from environmental
  variables `AWS_SECRET_ACCESS_KEY` or `AWS_SECRET_KEY` (in that order).
  If the environmental variables aren't set and Packer is running on
  an EC2 instance, Packer will check the instance metadata for IAM role
  keys.

* `source_ami` (string) - The source AMI whose root volume will be copied
  and provisioned on the currently running instance. This must be an
  EBS-backed AMI with a root volume snapshot that you have access to.

Optional:

* `chroot_mounts` (list of list of strings) - This is a list of additional
  devices to mount into the chroot environment. This configuration parameter
  requires some additional documentation which is in the "Chroot Mounts" section
  below. Please read that section for more information on how to use this.

* `copy_files` (list of strings) - Paths to files on the running EC2 instance
  that will be copied into the chroot environment prior to provisioning.
  This is useful, for example, to copy `/etc/resolv.conf` so that DNS lookups
  work.

* `device_path` (string) - The path to the device where the root volume
  of the source AMI will be attached. This defaults to "" (empty string),
  which forces Packer to find an open device automatically.

* `mount_command` (string) - The command to use to mount devices. This
  defaults to "mount". This may be useful to set if you want to set
  environmental variables or perhaps run it with `sudo` or so on.

* `mount_path` (string) - The path where the volume will be mounted. This is
  where the chroot environment will be. This defaults to
  `packer-amazon-chroot-volumes/{{.Device}}`. This is a configuration
  template where the `.Device` variable is replaced with the name of the
  device where the volume is attached.

* `unmount_command` (string) - Just like `mount_command`, except this is
  the command to unmount devices.

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
