---
description: |
    The amazon-ebs Packer builder is able to create Amazon AMIs backed by EBS
    volumes for use in EC2. For more information on the difference between
    EBS-backed instances and instance-store backed instances, see the storage for
    the root device section in the EC2 documentation.
layout: docs
page_title: 'Amazon EBS - Builders'
sidebar_current: 'docs-builders-amazon-ebsbacked'
---

# AMI Builder (EBS backed)

Type: `amazon-ebs`

The `amazon-ebs` Packer builder is able to create Amazon AMIs backed by EBS
volumes for use in [EC2](https://aws.amazon.com/ec2/). For more information on
the difference between EBS-backed instances and instance-store backed instances,
see the ["storage for the root device" section in the EC2
documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ComponentsAMIs.html#storage-for-the-root-device).

This builder builds an AMI by launching an EC2 instance from a source AMI,
provisioning that running machine, and then creating an AMI from that machine.
This is all done in your own AWS account. The builder will create temporary
keypairs, security group rules, etc. that provide it temporary access to the
instance while the image is being created. This simplifies configuration quite a
bit.

The builder does *not* manage AMIs. Once it creates an AMI and stores it in your
account, it is up to you to use, delete, etc. the AMI.

-&gt; **Note:** Temporary resources are, by default, all created with the prefix
`packer`. This can be useful if you want to restrict the security groups and
key pairs Packer is able to operate on.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `access_key` (string) - The access key used to communicate with AWS. [Learn
    how to set this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `ami_name` (string) - The name of the resulting AMI that will appear when
    managing AMIs in the AWS console or via APIs. This must be unique. To help
    make this unique, use a function like `timestamp` (see [template
    engine](/docs/templates/engine.html) for more info)

-   `instance_type` (string) - The EC2 instance type to use while building the
    AMI, such as `t2.small`.

-   `region` (string) - The name of the region, such as `us-east-1`, in which to
    launch the EC2 instance to create the AMI.

-   `secret_key` (string) - The secret key used to communicate with AWS. [Learn
    how to set this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `source_ami` (string) - The initial AMI used as a base for the newly
    created machine. `source_ami_filter` may be used instead to populate this
    automatically.

### Optional:

-   `ami_block_device_mappings` (array of block device mappings) - Add one or
    more [block device mappings](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/block-device-mapping-concepts.html)
    to the AMI. These will be attached when booting a new instance from your
    AMI. To add a block device during the Packer build see
    `launch_block_device_mappings` below. Your options here may vary depending
    on the type of VM you use. The block device mappings allow for the following
    configuration:

    -   `delete_on_termination` (boolean) - Indicates whether the EBS volume is
        deleted on instance termination. Default `false`. **NOTE**: If this
        value is not explicitly set to `true` and volumes are not cleaned up by
        an alternative method, additional volumes will accumulate after
        every build.

    -   `device_name` (string) - The device name exposed to the instance (for
        example, `/dev/sdh` or `xvdh`). Required when specifying `volume_size`.

    -   `encrypted` (boolean) - Indicates whether to encrypt the volume or not

    -   `iops` (integer) - The number of I/O operations per second (IOPS) that the
        volume supports. See the documentation on
        [IOPs](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_EbsBlockDevice.html)
        for more information

    -   `no_device` (boolean) - Suppresses the specified device included in the
        block device mapping of the AMI

    -   `snapshot_id` (string) - The ID of the snapshot

    -   `virtual_name` (string) - The virtual device name. See the documentation on
        [Block Device
        Mapping](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_BlockDeviceMapping.html)
        for more information

    -   `volume_size` (integer) - The size of the volume, in GiB. Required if not
        specifying a `snapshot_id`

    -   `volume_type` (string) - The volume type. `gp2` for General Purpose (SSD)
        volumes, `io1` for Provisioned IOPS (SSD) volumes, and `standard` for Magnetic
        volumes

-   `ami_description` (string) - The description to set for the
    resulting AMI(s). By default this description is empty. This is a
    [template engine](/docs/templates/engine.html)
    where the `SourceAMI` variable is replaced with the source AMI ID and
    `BuildRegion` variable is replaced with the value of `region`.

-   `ami_groups` (array of strings) - A list of groups that have access to
    launch the resulting AMI(s). By default no groups have permission to launch
    the AMI. `all` will make the AMI publicly accessible. AWS currently doesn't
    accept any value other than `all`.

-   `ami_product_codes` (array of strings) - A list of product codes to
    associate with the AMI. By default no product codes are associated with
    the AMI.

-   `ami_regions` (array of strings) - A list of regions to copy the AMI to.
    Tags and attributes are copied along with the AMI. AMI copying takes time
    depending on the size of the AMI, but will generally take many minutes.

-   `ami_users` (array of strings) - A list of account IDs that have access to
    launch the resulting AMI(s). By default no additional users other than the
    user creating the AMI has permissions to launch it.

-   `ami_virtualization_type` (string) - The type of virtualization for the AMI
    you are building. This option must match the supported virtualization
    type of `source_ami`. Can be `paravirtual` or `hvm`.

-   `associate_public_ip_address` (boolean) - If using a non-default VPC, public
    IP addresses are not provided by default. If this is toggled, your new
    instance will get a Public IP.

-   `availability_zone` (string) - Destination availability zone to launch
    instance in. Leave this empty to allow Amazon to auto-assign.

-   `custom_endpoint_ec2` (string) - this option is useful if you use
    another cloud provider that provide a compatible API with aws EC2,
    specify another endpoint like this "<https://ec2.another.endpoint>..com"

-   `disable_stop_instance` (boolean) - Packer normally stops the build instance
    after all provisioners have run. For Windows instances, it is sometimes
    desirable to [run Sysprep](http://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/ami-create-standard.html)
    which will stop the instance for you. If this is set to true, Packer *will not*
    stop the instance and will wait for you to stop it manually. You can do this
    with a [windows-shell provisioner](https://www.packer.io/docs/provisioners/windows-shell.html).

    ``` json
    {
      "type": "windows-shell",
      "inline": ["\"c:\\Program Files\\Amazon\\Ec2ConfigService\\ec2config.exe\" -sysprep"]
    }
    ```

-   `ebs_optimized` (boolean) - Mark instance as [EBS
    Optimized](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSOptimized.html).
    Default `false`.

-   `enhanced_networking` (boolean) - Enable enhanced
    networking (SriovNetSupport and ENA) on HVM-compatible AMIs. If true, add
    `ec2:ModifyInstanceAttribute` to your AWS IAM policy. Note: you must make
    sure enhanced networking is enabled on your instance. See [Amazon's
    documentation on enabling enhanced networking](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html#enabling_enhanced_networking)

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

-   `iam_instance_profile` (string) - The name of an [IAM instance
    profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/instance-profiles.html)
    to launch the EC2 instance with.

-   `launch_block_device_mappings` (array of block device mappings) - Add one or
    more block devices before the Packer build starts. These are not necessarily
    preserved when booting from the AMI built with Packer. See
    `ami_block_device_mappings`, above, for details.

-   `mfa_code` (string) - The MFA [TOTP](https://en.wikipedia.org/wiki/Time-based_One-time_Password_Algorithm)
    code. This should probably be a user variable since it changes all the time.

-   `profile` (string) - The profile to use in the shared credentials file for
    AWS. See Amazon's documentation on [specifying
    profiles](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-profiles)
    for more details.

-   `region_kms_key_ids` (map of strings) - a map of regions to copy the ami to,
    along with the custom kms key id to use for encryption for that region.
    Keys must match the regions provided in `ami_regions`. If you just want to
    encrypt using a default ID, you can stick with `kms_key_id` and `ami_regions`.
    If you want a region to be encrypted with that region's default key ID, you can
    use an empty string `""` instead of a key id in this map. (e.g. `"us-east-1": ""`)
    However, you cannot use default key IDs if you are using this in conjunction with
    `snapshot_users` -- in that situation you must use custom keys.

-   `run_tags` (object of key/value strings) - Tags to apply to the instance
    that is *launched* to create the AMI. These tags are *not* applied to the
    resulting AMI unless they're duplicated in `tags`. This is a
    [template engine](/docs/templates/engine.html)
    where the `SourceAMI` variable is replaced with the source AMI ID and
    `BuildRegion` variable is replaced with the value of `region`.

-   `run_volume_tags` (object of key/value strings) - Tags to apply to the volumes
    that are *launched* to create the AMI. These tags are *not* applied to the
    resulting AMI unless they're duplicated in `tags`. This is a
    [template engine](/docs/templates/engine.html)
    where the `SourceAMI` variable is replaced with the source AMI ID and
    `BuildRegion` variable is replaced with the value of `region`.

-   `security_group_id` (string) - The ID (*not* the name) of the security group
    to assign to the instance. By default this is not set and Packer will
    automatically create a new temporary security group to allow SSH access.
    Note that if this is specified, you must be sure the security group allows
    access to the `ssh_port` given below.

-   `security_group_ids` (array of strings) - A list of security groups as
    described above. Note that if this is specified, you must omit the
    `security_group_id`.

-   `shutdown_behavior` (string) - Automatically terminate instances on shutdown
    in case Packer exits ungracefully. Possible values are "stop" and "terminate",
    default is `stop`.

-   `skip_region_validation` (boolean) - Set to true if you want to skip
    validation of the region configuration option. Default `false`.

-   `snapshot_groups` (array of strings) - A list of groups that have access to
    create volumes from the snapshot(s). By default no groups have permission to create
    volumes form the snapshot(s). `all` will make the snapshot publicly accessible.

-   `snapshot_users` (array of strings) - A list of account IDs that have access to
    create volumes from the snapshot(s). By default no additional users other than the
    user creating the AMI has permissions to create volumes from the backing snapshot(s).

-   `snapshot_tags` (object of key/value strings) - Tags to apply to snapshot.
    They will override AMI tags if already applied to snapshot. This is a
    [template engine](/docs/templates/engine.html)
    where the `SourceAMI` variable is replaced with the source AMI ID and
    `BuildRegion` variable is replaced with the value of `region`.

-   `source_ami_filter` (object) - Filters used to populate the `source_ami` field.
    Example:

    ``` json
    {
      "source_ami_filter": {
        "filters": {
          "virtualization-type": "hvm",
          "name": "*ubuntu-xenial-16.04-amd64-server-*",
          "root-device-type": "ebs"
        },
        "owners": ["099720109477"],
        "most_recent": true
      }
    }
    ```

    This selects the most recent Ubuntu 16.04 HVM EBS AMI from Canonical.
    NOTE: This will fail unless *exactly* one AMI is returned. In the above
    example, `most_recent` will cause this to succeed by selecting the newest image.

    -   `filters` (map of strings) - filters used to select a `source_ami`.
        NOTE: This will fail unless *exactly* one AMI is returned.
        Any filter described in the docs for [DescribeImages](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeImages.html)
        is valid.

    -   `owners` (array of strings) - This scopes the AMIs to certain Amazon account IDs.
        This is helpful to limit the AMIs to a trusted third party, or to your own account.

    -   `most_recent` (bool) - Selects the newest created image when true.
        This is most useful for selecting a daily distro build.

-   `spot_price` (string) - The maximum hourly price to pay for a spot instance
    to create the AMI. Spot instances are a type of instance that EC2 starts
    when the current spot price is less than the maximum price you specify. Spot
    price will be updated based on available spot instance capacity and current
    spot instance requests. It may save you some costs. You can set this to
    `auto` for Packer to automatically discover the best spot price or to "0"
    to use an on demand instance (default).

-   `spot_price_auto_product` (string) - Required if `spot_price` is set
    to `auto`. This tells Packer what sort of AMI you're launching to find the
    best spot price. This must be one of: `Linux/UNIX`, `SUSE Linux`, `Windows`,
    `Linux/UNIX (Amazon VPC)`, `SUSE Linux (Amazon VPC)`, `Windows (Amazon VPC)`

-   `ssh_keypair_name` (string) - If specified, this is the key that will be
    used for SSH with the machine. The key must match a key pair name loaded
    up into Amazon EC2. By default, this is blank, and Packer will
    generate a temporary keypair unless
    [`ssh_password`](/docs/templates/communicator.html#ssh_password) is used.
    [`ssh_private_key_file`](/docs/templates/communicator.html#ssh_private_key_file)
    or `ssh_agent_auth` must be specified when `ssh_keypair_name` is utilized.

-   `ssh_agent_auth` (boolean) - If true, the local SSH agent will be used to
    authenticate connections to the source instance. No temporary keypair will
    be created, and the values of `ssh_password` and `ssh_private_key_file` will
    be ignored. To use this option with a key pair already configured in the source
    AMI, leave the `ssh_keypair_name` blank. To associate an existing key pair
    in AWS with the source instance, set the `ssh_keypair_name` field to the name
    of the key pair.

-   `ssh_private_ip` (boolean) - If true, then SSH will always use the private
    IP if available. Also works for WinRM.

-   `subnet_id` (string) - If using VPC, the ID of the subnet, such as
    `subnet-12345def`, where Packer will launch the EC2 instance. This field is
    required if you are using an non-default VPC.

-   `tags` (object of key/value strings) - Tags applied to the AMI and
    relevant snapshots. This is a
    [template engine](/docs/templates/engine.html)
    where the `SourceAMI` variable is replaced with the source AMI ID and
    `BuildRegion` variable is replaced with the value of `region`.

-   `temporary_key_pair_name` (string) - The name of the temporary key pair
    to generate. By default, Packer generates a name that looks like
    `packer_<UUID>`, where &lt;UUID&gt; is a 36 character unique identifier.

-   `token` (string) - The access token to use. This is different from the
    access key and secret key. If you're not sure what this is, then you
    probably don't need it. This will also be read from the `AWS_SESSION_TOKEN`
    environmental variable.

-   `user_data` (string) - User data to apply when launching the instance. Note
    that you need to be careful about escaping characters due to the templates
    being JSON. It is often more convenient to use `user_data_file`, instead.

-   `user_data_file` (string) - Path to a file that will be used for the user
    data when launching the instance.

-   `vpc_id` (string) - If launching into a VPC subnet, Packer needs the VPC ID
    in order to create a temporary security group within the VPC. Requires `subnet_id`
    to be set. If this field is left blank, Packer will try to get the VPC ID from the
    `subnet_id`.

-   `windows_password_timeout` (string) - The timeout for waiting for a Windows
    password for Windows instances. Defaults to 20 minutes. Example value: `10m`

## Basic Example

Here is a basic example. You will need to provide access keys, and may need to
change the AMI IDs according to what images exist at the time the template is run:

``` json
{
  "type": "amazon-ebs",
  "access_key": "YOUR KEY HERE",
  "secret_key": "YOUR SECRET KEY HERE",
  "region": "us-east-1",
  "source_ami": "ami-fce3c696",
  "instance_type": "t2.micro",
  "ssh_username": "ubuntu",
  "ami_name": "packer-quick-start {{timestamp}}"
}
```

-&gt; **Note:** Packer can also read the access key and secret access key from
environmental variables. See the configuration reference in the section above
for more information on what environmental variables Packer will look for.

Further information on locating AMI IDs and their relationship to instance types
and regions can be found in the AWS EC2 Documentation
[for Linux](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/finding-an-ami.html)
or [for Windows](http://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/finding-an-ami.html).

## Accessing the Instance to Debug

If you need to access the instance to debug for some reason, run the builder
with the `-debug` flag. In debug mode, the Amazon builder will save the private
key in the current directory and will output the DNS or IP information as well.
You can use this information to access the instance as it is running.

## AMI Block Device Mappings Example

Here is an example using the optional AMI block device mappings. Our
configuration of `launch_block_device_mappings` will expand the root volume
(`/dev/sda`) to 40gb during the build (up from the default of 8gb). With
`ami_block_device_mappings` AWS will attach additional volumes `/dev/sdb` and
`/dev/sdc` when we boot a new instance of our AMI.

``` json
{
  "type": "amazon-ebs",
  "access_key": "YOUR KEY HERE",
  "secret_key": "YOUR SECRET KEY HERE",
  "region": "us-east-1",
  "source_ami": "ami-fce3c696",
  "instance_type": "t2.micro",
  "ssh_username": "ubuntu",
  "ami_name": "packer-quick-start {{timestamp}}",
  "launch_block_device_mappings": [
    {
      "device_name": "/dev/sda1",
      "volume_size": 40,
      "volume_type": "gp2",
      "delete_on_termination": true
    }
  ],
  "ami_block_device_mappings": [
    {
      "device_name": "/dev/sdb",
      "virtual_name": "ephemeral0"
    },
    {
      "device_name": "/dev/sdc",
      "virtual_name": "ephemeral1"
    }
  ]
}
```

## Tag Example

Here is an example using the optional AMI tags. This will add the tags
`OS_Version` and `Release` to the finished AMI. As before, you will need to
provide your access keys, and may need to change the source AMI ID based on what
images exist when this template is run:

``` json
{
  "type": "amazon-ebs",
  "access_key": "YOUR KEY HERE",
  "secret_key": "YOUR SECRET KEY HERE",
  "region": "us-east-1",
  "source_ami": "ami-fce3c696",
  "instance_type": "t2.micro",
  "ssh_username": "ubuntu",
  "ami_name": "packer-quick-start {{timestamp}}",
  "tags": {
    "OS_Version": "Ubuntu",
    "Release": "Latest"
  }
}
```

-&gt; **Note:** Packer uses pre-built AMIs as the source for building images.
These source AMIs may include volumes that are not flagged to be destroyed on
termination of the instance building the new image. Packer will attempt to clean
up all residual volumes that are not designated by the user to remain after
termination. If you need to preserve those source volumes, you can overwrite the
termination setting by specifying `delete_on_termination=false` in the
`launch_block_device_mappings` block for the device.
