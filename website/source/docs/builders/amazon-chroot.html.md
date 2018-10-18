---
description: |
    The amazon-chroot Packer builder is able to create Amazon AMIs backed by an
    EBS volume as the root device. For more information on the difference between
    instance storage and EBS-backed instances, storage for the root device section
    in the EC2 documentation.
layout: docs
page_title: 'Amazon chroot - Builders'
sidebar_current: 'docs-builders-amazon-chroot'
---

# AMI Builder (chroot)

Type: `amazon-chroot`

The `amazon-chroot` Packer builder is able to create Amazon AMIs backed by an
EBS volume as the root device. For more information on the difference between
instance storage and EBS-backed instances, see the ["storage for the root
device" section in the EC2
documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ComponentsAMIs.html#storage-for-the-root-device).

The difference between this builder and the `amazon-ebs` builder is that this
builder is able to build an EBS-backed AMI without launching a new EC2 instance.
This can dramatically speed up AMI builds for organizations who need the extra
fast build.

~&gt; **This is an advanced builder** If you're just getting started with
Packer, we recommend starting with the [amazon-ebs
builder](/docs/builders/amazon-ebs.html), which is much easier to use.

The builder does *not* manage AMIs. Once it creates an AMI and stores it in your
account, it is up to you to use, delete, etc., the AMI.

## How Does it Work?

This builder works by creating a new EBS volume from an existing source AMI and
attaching it into an already-running EC2 instance. Once attached, a
[chroot](https://en.wikipedia.org/wiki/Chroot) is used to provision the system
within that volume. After provisioning, the volume is detached, snapshotted, and
an AMI is made.

Using this process, minutes can be shaved off the AMI creation process because a
new EC2 instance doesn't need to be launched.

There are some restrictions, however. The host EC2 instance where the volume is
attached to must be a similar system (generally the same OS version, kernel
versions, etc.) as the AMI being built. Additionally, this process is much more
expensive because the EC2 instance must be kept running persistently in order to
build AMIs, whereas the other AMI builders start instances on-demand to build
AMIs as needed.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

### Required:

-   `access_key` (string) - The access key used to communicate with AWS. [Learn
    how to set this](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `ami_name` (string) - The name of the resulting AMI that will appear when
    managing AMIs in the AWS console or via APIs. This must be unique. To help
    make this unique, use a function like `timestamp` (see [template
    engine](/docs/templates/engine.html) for more info).

-   `secret_key` (string) - The secret key used to communicate with AWS. [Learn
    how to set this](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `source_ami` (string) - The source AMI whose root volume will be copied and
    provisioned on the currently running instance. This must be an EBS-backed AMI
    with a root volume snapshot that you have access to. Note: this is not used
    when `from_scratch` is set to `true`.

### Optional:

-   `ami_description` (string) - The description to set for the
    resulting AMI(s). By default this description is empty. This is a
    [template engine](/docs/templates/engine.html),
    see [Build template data](#build-template-data) for more information.

-   `ami_groups` (array of strings) - A list of groups that have access to
    launch the resulting AMI(s). By default no groups have permission to launch
    the AMI. `all` will make the AMI publicly accessible.

-   `ami_product_codes` (array of strings) - A list of product codes to
    associate with the AMI. By default no product codes are associated with
    the AMI.

-   `ami_regions` (array of strings) - A list of regions to copy the AMI to.
    Tags and attributes are copied along with the AMI. AMI copying takes time
    depending on the size of the AMI, but will generally take many minutes.

-   `ami_users` (array of strings) - A list of account IDs that have access to
    launch the resulting AMI(s). By default no additional users other than the user creating the AMI has permissions to launch it.

-   `ami_virtualization_type` (string) - The type of virtualization for the AMI
    you are building. This option is required to register HVM images. Can be
    `paravirtual` (default) or `hvm`.

-   `chroot_mounts` (array of array of strings) - This is a list of devices
    to mount into the chroot environment. This configuration parameter
    requires some additional documentation which is in the [Chroot Mounts](#Chroot Mounts)
    section. Please read that section for more information on how to
    use this.

-   `command_wrapper` (string) - How to run shell commands. This defaults to
    `{{.Command}}`. This may be useful to set if you want to set environmental
    variables or perhaps run it with `sudo` or so on. This is a configuration
    template where the `.Command` variable is replaced with the command to
    be run. Defaults to `{{.Command}}`.

-   `copy_files` (array of strings) - Paths to files on the running EC2 instance
    that will be copied into the chroot environment prior to provisioning. Defaults
    to `/etc/resolv.conf` so that DNS lookups work. Pass an empty list to skip
    copying `/etc/resolv.conf`. You may need to do this if you're building
    an image that uses systemd.

-   `custom_endpoint_ec2` (string) - This option is useful if you use a cloud
    provider whose API is compatible with aws EC2. Specify another endpoint
    like this `https://ec2.custom.endpoint.com`.

-   `decode_authorization_messages` (boolean) - Enable automatic decoding of any
    encoded authorization (error) messages using the `sts:DecodeAuthorizationMessage` API.
    Note: requires that the effective user/role have permissions to `sts:DecodeAuthorizationMessage`
    on resource `*`. Default `false`.

-   `device_path` (string) - The path to the device where the root volume of the
    source AMI will be attached. This defaults to "" (empty string), which
    forces Packer to find an open device automatically.

-   `ena_support` (boolean) - Enable enhanced networking (ENA but not SriovNetSupport)
    on HVM-compatible AMIs. If `true`, add `ec2:ModifyInstanceAttribute` to your AWS IAM policy.
    Note: you must make sure enhanced networking is enabled on your instance. See [Amazon's
    documentation on enabling enhanced networking](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html#enabling_enhanced_networking). Default `false`.

-   `force_deregister` (boolean) - Force Packer to first deregister an existing
    AMI if one with the same name already exists. Default `false`.

-   `force_delete_snapshot` (boolean) - Force Packer to delete snapshots associated with
    AMIs, which have been deregistered by `force_deregister`. Default `false`.

-   `encrypt_boot` (boolean) - Instruct packer to automatically create a copy of the
    AMI with an encrypted boot volume (discarding the initial unencrypted AMI in the
    process). Packer will always run this operation, even if the base
    AMI has an encrypted boot volume to start with. Default `false`.

-   `kms_key_id` (string) - The ID of the KMS key to use for boot volume encryption.
    This only applies to the main `region`, other regions where the AMI will be copied
    will be encrypted by the default EBS KMS key.

-   `from_scratch` (boolean) - Build a new volume instead of starting from an
    existing AMI root volume snapshot. Default `false`. If `true`, `source_ami` is
    no longer used and the following options become required:
    `ami_virtualization_type`, `pre_mount_commands` and `root_volume_size`. The
    below options are also required in this mode only:

-   `ami_block_device_mappings` (array of block device mappings) - Add one or
    more [block device mappings](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html)
    to the AMI. These will be attached when booting a new instance from your
    AMI. Your options here may vary depending on the type of VM you use. The
    block device mappings allow for the following configuration:

    -   `delete_on_termination` (boolean) - Indicates whether the EBS volume is
        deleted on instance termination. Default `false`. **NOTE**: If this
        value is not explicitly set to `true` and volumes are not cleaned up by
        an alternative method, additional volumes will accumulate after
        every build.

    -   `device_name` (string) - The device name exposed to the instance (for
        example, `/dev/sdh` or `xvdh`). Required for every device in the
        block device mapping.

    -   `encrypted` (boolean) - Indicates whether or not to encrypt the volume.

    -   `kms_key_id` (string) - The ARN for the KMS encryption key. When
        specifying `kms_key_id`, `encrypted` needs to be set to `true`.

    -   `iops` (number) - The number of I/O operations per second (IOPS) that the
        volume supports. See the documentation on
        [IOPS](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_EbsBlockDevice.html)
        for more information.

    -   `no_device` (boolean) - Suppresses the specified device included in the
        block device mapping of the AMI.

    -   `snapshot_id` (string) - The ID of the snapshot.

    -   `virtual_name` (string) - The virtual device name. See the documentation on
        [Block Device
        Mapping](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_BlockDeviceMapping.html)
        for more information.

    -   `volume_size` (number) - The size of the volume, in GiB. Required if not
        specifying a `snapshot_id`.

    -   `volume_type` (string) - The volume type. `gp2` for General Purpose (SSD)
        volumes, `io1` for Provisioned IOPS (SSD) volumes, and `standard` for Magnetic
        volumes

-   `region_kms_key_ids` (map of strings) - a map of regions to copy the ami to,
    along with the custom kms key id to use for encryption for that region.
    Keys must match the regions provided in `ami_regions`. If you just want to
    encrypt using a default ID, you can stick with `kms_key_id` and `ami_regions`.
    If you want a region to be encrypted with that region's default key ID, you can
    use an empty string `""` instead of a key id in this map. (e.g. `"us-east-1": ""`)
    However, you cannot use default key IDs if you are using this in conjunction with
    `snapshot_users` -- in that situation you must use custom keys.

-   `root_device_name` (string) - The root device name. For example, `xvda`.

-   `mfa_code` (string) - The MFA [TOTP](https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm)
    code. This should probably be a user variable since it changes all the time.

-   `mount_path` (string) - The path where the volume will be mounted. This is
    where the chroot environment will be. This defaults to
    `/mnt/packer-amazon-chroot-volumes/{{.Device}}`. This is a configuration template
    where the `.Device` variable is replaced with the name of the device where
    the volume is attached.

-   `mount_partition` (string) - The partition number containing the
    / partition. By default this is the first partition of the volume, (for
    example, `xvda1`) but you can designate the entire block device by setting
    `"mount_partition": "0"` in your config, which will mount `xvda` instead.

-   `mount_options` (array of strings) - Options to supply the `mount` command
    when mounting devices. Each option will be prefixed with `-o` and supplied
    to the `mount` command ran by Packer. Because this command is ran in a
    shell, user discretion is advised. See [this manual page for the mount
    command](http://linuxcommand.org/man_pages/mount8.html) for valid file
    system specific options.

-   `nvme_device_path` (string) - When we call the mount command (by default
    `mount -o device dir`), the string provided in `nvme_mount_path` will
    replace `device` in that command. When this option is not set, `device` in
    that command will be something like `/dev/sdf1`, mirroring the attached
    device name. This assumption works for most instances but will fail with c5
    and m5 instances. In order to use the chroot builder with c5 and m5
    instances, you must manually set `nvme_device_path` and `device_path`.

-   `pre_mount_commands` (array of strings) - A series of commands to execute
    after attaching the root volume and before mounting the chroot. This is not
    required unless using `from_scratch`. If so, this should include any
    partitioning and filesystem creation commands. The path to the device is
    provided by `{{.Device}}`.

-   `profile` (string) - The profile to use in the shared credentials file for
    AWS. See Amazon's documentation on [specifying
    profiles](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-profiles)
    for more details.

-   `post_mount_commands` (array of strings) - As `pre_mount_commands`, but the
    commands are executed after mounting the root device and before the extra
    mount and copy steps. The device and mount path are provided by
    `{{.Device}}` and `{{.MountPath}}`.

-   `root_volume_size` (number) - The size of the root volume in GB for the
    chroot environment and the resulting AMI. Default size is the snapshot size
    of the `source_ami` unless `from_scratch` is `true`, in which case
    this field must be defined.

-   `root_volume_type` (string) - The type of EBS volume for the chroot environment
     and resulting AMI. The default value is the type of the `source_ami`, unless
    `from_scratch` is `true`, in which case the default value is `gp2`. You can only
     specify `io1` if building based on top of a `source_ami` which is also `io1`.

-   `root_volume_tags` (object of key/value strings) - Tags to apply to the volumes
    that are *launched*. This is a
    [template engine](/docs/templates/engine.html),
    see [Build template data](#build-template-data) for more information.

-   `skip_region_validation` (boolean) - Set to `true` if you want to skip
    validation of the `ami_regions` configuration option. Default `false`.

-   `snapshot_tags` (object of key/value strings) - Tags to apply to snapshot.
    They will override AMI tags if already applied to snapshot. This is a
    [template engine](/docs/templates/engine.html),
    see [Build template data](#build-template-data) for more information.

-   `snapshot_groups` (array of strings) - A list of groups that have access to
    create volumes from the snapshot(s). By default no groups have permission to create
    volumes from the snapshot(s). `all` will make the snapshot publicly accessible.

-   `snapshot_users` (array of strings) - A list of account IDs that have access to
    create volumes from the snapshot(s). By default no additional users other than the
    user creating the AMI has permissions to create volumes from the backing snapshot(s).

-   `source_ami_filter` (object) - Filters used to populate the `source_ami` field.
    Example:

    ``` json
    "source_ami_filter": {
      "filters": {
        "virtualization-type": "hvm",
        "name": "ubuntu/images/*ubuntu-xenial-16.04-amd64-server-*",
        "root-device-type": "ebs"
      },
      "owners": ["099720109477"],
      "most_recent": true
    }
    ```

    This selects the most recent Ubuntu 16.04 HVM EBS AMI from Canonical.
    NOTE: This will fail unless *exactly* one AMI is returned. In the above
    example, `most_recent` will cause this to succeed by selecting the newest image.

    -   `filters` (map of strings) - filters used to select a `source_ami`.
        NOTE: This will fail unless *exactly* one AMI is returned.
        Any filter described in the docs for [DescribeImages](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeImages.html)
        is valid.

    -   `owners` (array of strings) - Filters the images by their owner. You may
        specify one or more AWS account IDs, "self" (which will use the account
        whose credentials you are using to run Packer), or an AWS owner alias:
        for example, "amazon", "aws-marketplace", or "microsoft".
        This option is required for security reasons.

    -   `most_recent` (boolean) - Selects the newest created image when `true`.
        This is most useful for selecting a daily distro build.

    You may set this in place of `source_ami` or in conjunction with it. If you
    set this in conjunction with `source_ami`, the `source_ami` will be added to
    the filter. The provided `source_ami` must meet all of the filtering criteria
    provided in `source_ami_filter`; this pins the AMI returned by the filter,
    but will cause Packer to fail if the `source_ami` does not exist.

-   `sriov_support` (boolean) - Enable enhanced networking (SriovNetSupport but not ENA)
    on HVM-compatible AMIs. If `true`, add `ec2:ModifyInstanceAttribute` to your AWS IAM
    policy. Note: you must make sure enhanced networking is enabled on your instance. See [Amazon's
    documentation on enabling enhanced networking](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html#enabling_enhanced_networking).
    Default `false`.

-   `tags` (object of key/value strings) - Tags applied to the AMI. This is a
    [template engine](/docs/templates/engine.html),
    see [Build template data](#build-template-data) for more information.

## Basic Example

Here is a basic example. It is completely valid except for the access keys:

``` json
{
  "type": "amazon-chroot",
  "access_key": "YOUR KEY HERE",
  "secret_key": "YOUR SECRET KEY HERE",
  "source_ami": "ami-e81d5881",
  "ami_name": "packer-amazon-chroot {{timestamp}}"
}
```

## Chroot Mounts

The `chroot_mounts` configuration can be used to mount specific devices within
the chroot. By default, the following additional mounts are added into the
chroot by Packer:

-   `/proc` (proc)
-   `/sys` (sysfs)
-   `/dev` (bind to real `/dev`)
-   `/dev/pts` (devpts)
-   `/proc/sys/fs/binfmt_misc` (binfmt\_misc)

These default mounts are usually good enough for anyone and are sane defaults.
However, if you want to change or add the mount points, you may using the
`chroot_mounts` configuration. Here is an example configuration which only
mounts `/proc` and `/dev`:

``` json
{
  "chroot_mounts": [
    ["proc", "proc", "/proc"],
    ["bind", "/dev", "/dev"]
  ]
}
```

`chroot_mounts` is a list of a 3-tuples of strings. The three components of the
3-tuple, in order, are:

-   The filesystem type. If this is "bind", then Packer will properly bind the
    filesystem to another mount point.

-   The source device.

-   The mount directory.

## Parallelism

A quick note on parallelism: it is perfectly safe to run multiple *separate*
Packer processes with the `amazon-chroot` builder on the same EC2 instance. In
fact, this is recommended as a way to push the most performance out of your AMI
builds.

Packer properly obtains a process lock for the parallelism-sensitive parts of
its internals such as finding an available device.

## Gotchas

### Unmounting the Filesystem
One of the difficulties with using the chroot builder is that your provisioning
scripts must not leave any processes running or packer will be unable to unmount
the filesystem.

For debian based distributions you can setup a
[policy-rc.d](http://people.debian.org/~hmh/invokerc.d-policyrc.d-specification.txt)
file which will prevent packages installed by your provisioners from starting
services:

``` json
{
  "type": "shell",
  "inline": [
    "echo '#!/bin/sh' > /usr/sbin/policy-rc.d",
    "echo 'exit 101' >> /usr/sbin/policy-rc.d",
    "chmod a+x /usr/sbin/policy-rc.d"
  ]
},

// ...

{
  "type": "shell",
  "inline": [
    "rm -f /usr/sbin/policy-rc.d"
  ]
}
```

### Using Instances with NVMe block devices.
In C5, C5d, M5, and i3.metal instances, EBS volumes are exposed as NVMe block
devices [reference](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/nvme-ebs-volumes.html).
In order to correctly mount these devices, you have to do some extra legwork,
involving the `nvme_device_path` option above. Read that for more information.

A working example for mounting an NVMe device is below:

```
{
  "variables": {
    "region" : "us-east-2"
  },
  "builders": [
    {
      "type": "amazon-chroot",
      "region": "{{user `region`}}",
      "source_ami_filter": {
        "filters": {
        "virtualization-type": "hvm",
        "name": "amzn-ami-hvm-*",
        "root-device-type": "ebs"
        },
        "owners": ["137112412989"],
        "most_recent": true
      },
      "ena_support": true,
      "ami_name": "amazon-chroot-test-{{timestamp}}",
      "nvme_device_path": "/dev/nvme1n1p",
      "device_path": "/dev/sdf"
    }
  ],

  "provisioners": [
    {
      "type": "shell",
      "inline": ["echo Test > /tmp/test.txt"]
    }
  ]
}
```

Note that in the `nvme_device_path` you must end with the `p`; if you try to
define the partition in this path (e.g. `nvme_device_path`: `/dev/nvme1n1p1`)
and haven't also set the `"mount_partition": 0`, a `1` will be appended to the
`nvme_device_path` and Packer will fail.

## Building From Scratch

This example demonstrates the essentials of building an image from scratch. A
15G gp2 (SSD) device is created (overriding the default of standard/magnetic).
The device setup commands partition the device with one partition for use as an
HVM image and format it ext4. This builder block should be followed by
provisioning commands to install the os and bootloader.

``` json
{
  "type": "amazon-chroot",
  "ami_name": "packer-from-scratch {{timestamp}}",
  "from_scratch": true,
  "ami_virtualization_type": "hvm",
  "pre_mount_commands": [
    "parted {{.Device}} mklabel msdos mkpart primary 1M 100% set 1 boot on print",
    "mkfs.ext4 {{.Device}}1"
  ],
  "root_volume_size": 15,
  "root_device_name": "xvda",
  "ami_block_device_mappings": [
    {
      "device_name": "xvda",
      "delete_on_termination": true,
      "volume_type": "gp2"
    }
  ]
}
```

## Build template data

The available variables are:

- `BuildRegion` - The region (for example `eu-central-1`) where Packer is building the AMI.
- `SourceAMI` - The source AMI ID (for example `ami-a2412fcd`) used to build the AMI.
- `SourceAMIName` - The source AMI Name (for example `ubuntu/images/ebs-ssd/ubuntu-xenial-16.04-amd64-server-20180306`) used to build the AMI.
- `SourceAMITags` - The source AMI Tags, as a `map[string]string` object.
