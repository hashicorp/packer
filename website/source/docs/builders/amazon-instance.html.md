---
description: |
    The amazon-instance Packer builder is able to create Amazon AMIs backed by
    instance storage as the root device. For more information on the difference
    between instance storage and EBS-backed instances, see the storage for the
    root device section in the EC2 documentation.
layout: docs
page_title: 'Amazon instance-store - Builders'
sidebar_current: 'docs-builders-amazon-instance'
---

# AMI Builder (instance-store)

Type: `amazon-instance`

The `amazon-instance` Packer builder is able to create Amazon AMIs backed by
instance storage as the root device. For more information on the difference
between instance storage and EBS-backed instances, see the ["storage for the
root device" section in the EC2
documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ComponentsAMIs.html#storage-for-the-root-device).

This builder builds an AMI by launching an EC2 instance from an existing
instance-storage backed AMI, provisioning that running machine, and then
bundling and creating a new AMI from that machine. This is all done in your own
AWS account. This builder will create temporary key pairs, security group rules,
etc. that provide it temporary access to the instance while the image is being
created. This simplifies configuration quite a bit.

This builder does *not* manage AMIs. Once it creates an AMI and stores it in
your account, it is up to you to use, delete, etc. the AMI.

-&gt; **Note:** Temporary resources are, by default, all created with the prefix
`packer`. This can be useful if you want to restrict the security groups and
key pairs packer is able to operate on.

-&gt; **Note:** This builder requires that the [Amazon EC2 AMI
Tools](https://aws.amazon.com/developertools/368) are installed onto the
machine. This can be done within a provisioner, but must be done before the
builder finishes running.

~&gt; Instance builds are not supported for Windows. Use [`amazon-ebs`](amazon-ebs.html) instead.

## Configuration Reference

There are many configuration options available for this builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `access_key` (string) - The access key used to communicate with AWS. [Learn
    how to set this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `account_id` (string) - Your AWS account ID. This is required for bundling
    the AMI. This is *not the same* as the access key. You can find your account
    ID in the security credentials page of your AWS account.

-   `ami_name` (string) - The name of the resulting AMI that will appear when
    managing AMIs in the AWS console or via APIs. This must be unique. To help
    make this unique, use a function like `timestamp` (see [configuration
    templates](/docs/templates/engine.html) for more info)

-   `instance_type` (string) - The EC2 instance type to use while building the
    AMI, such as `m1.small`.

-   `region` (string) - The name of the region, such as `us-east-1`, in which to
    launch the EC2 instance to create the AMI.

-   `s3_bucket` (string) - The name of the S3 bucket to upload the AMI. This
    bucket will be created if it doesn't exist.

-   `secret_key` (string) - The secret key used to communicate with AWS. [Learn
    how to set this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `source_ami` (string) - The initial AMI used as a base for the newly
    created machine.

-   `x509_cert_path` (string) - The local path to a valid X509 certificate for
    your AWS account. This is used for bundling the AMI. This X509 certificate
    must be registered with your account from the security credentials page in
    the AWS console.

-   `x509_key_path` (string) - The local path to the private key for the X509
    certificate specified by `x509_cert_path`. This is used for bundling
    the AMI.

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
        example, `/dev/sdh` or `xvdh`). Required for every device in the
        block device mapping.

    -   `encrypted` (boolean) - Indicates whether to encrypt the volume or not

    -   `iops` (number) - The number of I/O operations per second (IOPS) that the
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

    -   `volume_size` (number) - The size of the volume, in GiB. Required if not
        specifying a `snapshot_id`

    -   `volume_type` (string) - The volume type. `gp2` for General Purpose (SSD)
        volumes, `io1` for Provisioned IOPS (SSD) volumes, and `standard` for Magnetic
        volumes

-   `ami_description` (string) - The description to set for the
    resulting AMI(s). By default this description is empty. This is a
    [template engine](/docs/templates/engine.html),
    see [Build template data](#build-template-data) for more information.

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
    you are building. This option is required to register HVM images. Can be
    `paravirtual` (default) or `hvm`.

-   `associate_public_ip_address` (boolean) - If using a non-default VPC, public
    IP addresses are not provided by default. If this is toggled, your new
    instance will get a Public IP.

-   `availability_zone` (string) - Destination availability zone to launch
    instance in. Leave this empty to allow Amazon to auto-assign.

-   `block_duration_minutes` (int64) - Requires `spot_price` to
    be set. The required duration for the Spot Instances (also known as Spot blocks).
    This value must be a multiple of 60 (60, 120, 180, 240, 300, or 360).
    You can't specify an Availability Zone group or a launch group if you specify a duration.

-   `bundle_destination` (string) - The directory on the running instance where
    the bundled AMI will be saved prior to uploading. By default this is `/tmp`.
    This directory must exist and be writable.

-   `bundle_prefix` (string) - The prefix for files created from bundling the
    root volume. By default this is `image-{{timestamp}}`. The `timestamp`
    variable should be used to make sure this is unique, otherwise it can
    collide with other created AMIs by Packer in your account.

-   `bundle_upload_command` (string) - The command to use to upload the
    bundled volume. See the "custom bundle commands" section below for
    more information.

-   `bundle_vol_command` (string) - The command to use to bundle the volume. See
    the "custom bundle commands" section below for more information.

-   `custom_endpoint_ec2` (string) - This option is useful if you use a cloud
    provider whose API is compatible with aws EC2. Specify another endpoint
    like this `https://ec2.custom.endpoint.com`.

-   `decode_authorization_messages` (boolean) - Enable automatic decoding of any
    encoded authorization (error) messages using the `sts:DecodeAuthorizationMessage` API.
    Note: requires that the effective user/role have permissions to `sts:DecodeAuthorizationMessage`
    on resource `*`. Default `false`.

-   `ebs_optimized` (boolean) - Mark instance as [EBS
    Optimized](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSOptimized.html).
    Default `false`.

-   `ena_support` (boolean) - Enable enhanced networking (ENA but not SriovNetSupport)
    on HVM-compatible AMIs. If true, add `ec2:ModifyInstanceAttribute` to your AWS IAM policy.
    Note: you must make sure enhanced networking is enabled on your instance. See [Amazon's
    documentation on enabling enhanced networking](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html#enabling_enhanced_networking). Default `false`.

-   `enable_t2_unlimited` (boolean) - Enabling T2 Unlimited allows the source
    instance to burst additional CPU beyond its available [CPU Credits]
    (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/t2-credits-baseline-concepts.html)
    for as long as the demand exists.
    This is in contrast to the standard configuration that only allows an
    instance to consume up to its available CPU Credits.
    See the AWS documentation for [T2 Unlimited]
    (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/t2-unlimited.html)
    and the 'T2 Unlimited Pricing' section of the [Amazon EC2 On-Demand
    Pricing](https://aws.amazon.com/ec2/pricing/on-demand/) document for more
    information.
    By default this option is disabled and Packer will set up a [T2
    Standard](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/t2-std.html)
    instance instead.

    To use T2 Unlimited you must use a T2 instance type e.g. t2.micro.
    Additionally, T2 Unlimited cannot be used in conjunction with Spot
    Instances e.g. when the `spot_price` option has been configured.
    Attempting to do so will cause an error.

    !&gt; **Warning!** Additional costs may be incurred by enabling T2
    Unlimited - even for instances that would usually qualify for the
    [AWS Free Tier](https://aws.amazon.com/free/).

-   `force_deregister` (boolean) - Force Packer to first deregister an existing
    AMI if one with the same name already exists. Defaults to `false`.

-   `force_delete_snapshot` (boolean) - Force Packer to delete snapshots associated with
    AMIs, which have been deregistered by `force_deregister`. Defaults to `false`.

-   `iam_instance_profile` (string) - The name of an [IAM instance
    profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/instance-profiles.html)
    to launch the EC2 instance with.

-   `launch_block_device_mappings` (array of block device mappings) - Add one
    or more block devices before the Packer build starts. If you add instance
    store volumes or EBS volumes in addition to the root device volume, the
    created AMI will contain block device mapping information for those
    volumes. Amazon creates snapshots of the source instance's root volume and
    any other EBS volumes described here. When you launch an instance from this
    new AMI, the instance automatically launches with these additional volumes,
    and will restore them from snapshots taken from the source instance.

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
    [template engine](/docs/templates/engine.html),
    see [Build template data](#build-template-data) for more information.

-   `security_group_id` (string) - The ID (*not* the name) of the security group
    to assign to the instance. By default this is not set and Packer will
    automatically create a new temporary security group to allow SSH access.
    Note that if this is specified, you must be sure the security group allows
    access to the `ssh_port` given below.

-   `security_group_ids` (array of strings) - A list of security groups as
    described above. Note that if this is specified, you must omit the
    `security_group_id`.

-   `security_group_filter` (object) - Filters used to populate the `security_group_ids` field.
    Example:

    ``` json
    {
      "security_group_filter": {
        "filters": {
          "tag:Class": "packer"
        }
      }
    }
    ```

    This selects the SG's with tag `Class` with the value `packer`.

    -   `filters` (map of strings) - filters used to select a `security_group_ids`.
        Any filter described in the docs for [DescribeSecurityGroups](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSecurityGroups.html)
        is valid.

    `security_group_ids` take precedence over this.

-   `temporary_security_group_source_cidr` (string) - An IPv4 CIDR block to be authorized
    access to the instance, when packer is creating a temporary security group.
    The default is `0.0.0.0/0` (i.e., allow any IPv4 source). This is only used
    when `security_group_id` or `security_group_ids` is not specified.

-   `skip_region_validation` (boolean) - Set to true if you want to skip
    validation of the region configuration option. Defaults to `false`.

-   `snapshot_groups` (array of strings) - A list of groups that have access to
    create volumes from the snapshot(s). By default no groups have permission to create
    volumes form the snapshot(s). `all` will make the snapshot publicly accessible.

-   `snapshot_users` (array of strings) - A list of account IDs that have access to
    create volumes from the snapshot(s). By default no additional users other than the
    user creating the AMI has permissions to create volumes from the backing snapshot(s).

-   `source_ami_filter` (object) - Filters used to populate the `source_ami` field.
    Example:

    ``` json
    {
      "source_ami_filter": {
        "filters": {
          "virtualization-type": "hvm",
          "name": "ubuntu/images/*ubuntu-xenial-16.04-amd64-server-*",
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

    -   `owners` (array of strings) - Filters the images by their owner. You may
        specify one or more AWS account IDs, "self" (which will use the account
        whose credentials you are using to run Packer), or an AWS owner alias:
        for example, "amazon", "aws-marketplace", or "microsoft".
        This option is required for security reasons.

    -   `most_recent` (boolean) - Selects the newest created image when true.
        This is most useful for selecting a daily distro build.

    You may set this in place of `source_ami` or in conjunction with it. If you
    set this in conjunction with `source_ami`, the `source_ami` will be added to
    the filter. The provided `source_ami` must meet all of the filtering criteria
    provided in `source_ami_filter`; this pins the AMI returned by the filter,
    but will cause Packer to fail if the `source_ami` does not exist.

-   `snapshot_tags` (object of key/value strings) - Tags to apply to snapshot.
    They will override AMI tags if already applied to snapshot.

-   `spot_price` (string) - The maximum hourly price to launch a spot instance
    to create the AMI. It is a type of instances that EC2 starts when the
    maximum price that you specify exceeds the current spot price. Spot price
    will be updated based on available spot instance capacity and current spot
    Instance requests. It may save you some costs. You can set this to `auto`
    for Packer to automatically discover the best spot price or to `0` to use
    an on-demand instance (default).

-   `spot_price_auto_product` (string) - Required if `spot_price` is set
    to `auto`. This tells Packer what sort of AMI you're launching to find the
    best spot price. This must be one of: `Linux/UNIX`, `SUSE Linux`, `Windows`,
    `Linux/UNIX (Amazon VPC)`, `SUSE Linux (Amazon VPC)`, `Windows (Amazon VPC)`

-   `spot_tags` (object of key/value strings) - Requires `spot_price` to
    be set. This tells Packer to apply tags to the spot request that is
    issued.

-   `sriov_support` (boolean) - Enable enhanced networking (SriovNetSupport but not ENA)
    on HVM-compatible AMIs. If true, add `ec2:ModifyInstanceAttribute` to your AWS IAM
    policy. Note: you must make sure enhanced networking is enabled on your instance. See [Amazon's
    documentation on enabling enhanced networking](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/enhanced-networking.html#enabling_enhanced_networking).
    Default `false`.

-   `ssh_keypair_name` (string) - If specified, this is the key that will be
    used for SSH with the machine. The key must match a key pair name loaded
    up into Amazon EC2. By default, this is blank, and Packer will
    generate a temporary key pair unless
    [`ssh_password`](/docs/templates/communicator.html#ssh_password) is used.
    [`ssh_private_key_file`](/docs/templates/communicator.html#ssh_private_key_file)
    or `ssh_agent_auth` must be specified when `ssh_keypair_name` is utilized.

-   `ssh_agent_auth` (boolean) - If true, the local SSH agent will be used to
    authenticate connections to the source instance. No temporary key pair will
    be created, and the values of `ssh_password` and `ssh_private_key_file` will
    be ignored. To use this option with a key pair already configured in the source
    AMI, leave the `ssh_keypair_name` blank. To associate an existing key pair
    in AWS with the source instance, set the `ssh_keypair_name` field to the name
    of the key pair.

-   `ssh_private_ip` (boolean) - No longer supported. See
    [`ssh_interface`](#ssh_interface). A fixer exists to migrate.

-   `ssh_interface` (string) - One of `public_ip`, `private_ip`,
    `public_dns` or `private_dns`. If set, either the public IP address,
    private IP address, public DNS name or private DNS name will used as the host for SSH.
    The default behaviour if inside a VPC is to use the public IP address if available,
    otherwise the private IP address will be used. If not in a VPC the public DNS name
    will be used. Also works for WinRM.

    Where Packer is configured for an outbound proxy but WinRM traffic should be direct,
    `ssh_interface` must be set to `private_dns` and `<region>.compute.internal` included
    in the `NO_PROXY` environment variable.

-   `subnet_id` (string) - If using VPC, the ID of the subnet, such as
    `subnet-12345def`, where Packer will launch the EC2 instance. This field is
    required if you are using an non-default VPC.

-   `subnet_filter` (object) - Filters used to populate the `subnet_id` field.
    Example:

    ``` json
    {
      "subnet_filter": {
        "filters": {
          "tag:Class": "build"
        },
        "most_free": true,
        "random": false
      }
    }
    ```

    This selects the Subnet with tag `Class` with the value `build`,  which has
    the most free IP addresses.
    NOTE: This will fail unless *exactly* one Subnet is returned. By using
    `most_free` or `random` one will be selected from those matching the filter.

    -   `filters` (map of strings) - filters used to select a `subnet_id`.
        NOTE: This will fail unless *exactly* one Subnet is returned.
        Any filter described in the docs for [DescribeSubnets](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSubnets.html)
        is valid.

    -   `most_free` (boolean) - The Subnet with the most free IPv4 addresses
        will be used if multiple Subnets matches the filter.

    -   `random` (boolean) - A random Subnet will be used if multiple Subnets
        matches the filter. `most_free` have precendence over this.

    `subnet_id` take precedence over this.

-   `tags` (object of key/value strings) - Tags applied to the AMI. This is a
    [template engine](/docs/templates/engine.html),
    see [Build template data](#build-template-data) for more information.

-   `temporary_key_pair_name` (string) - The name of the temporary key pair
    to generate. By default, Packer generates a name that looks like
    `packer_<UUID>`, where &lt;UUID&gt; is a 36 character unique identifier.

-   `user_data` (string) - User data to apply when launching the instance. Note
    that you need to be careful about escaping characters due to the templates
    being JSON. It is often more convenient to use `user_data_file`, instead.

-   `user_data_file` (string) - Path to a file that will be used for the user
    data when launching the instance.

-   `vpc_id` (string) - If launching into a VPC subnet, Packer needs the VPC ID
    in order to create a temporary security group within the VPC. Requires `subnet_id`
    to be set. If this field is left blank, Packer will try to get the VPC ID from the
    `subnet_id`.

-   `vpc_filter` (object) - Filters used to populate the `vpc_id` field.
    Example:

    ``` json
    {
      "vpc_filter": {
        "filters": {
          "tag:Class": "build",
          "isDefault": "false",
          "cidr": "/24"
        }
      }
    }
    ```

    This selects the VPC with tag `Class` with the value `build`,  which is not the
    default VPC, and have a IPv4 CIDR block of `/24`.
    NOTE: This will fail unless *exactly* one VPC is returned.

    -   `filters` (map of strings) - filters used to select a `vpc_id`.
        NOTE: This will fail unless *exactly* one VPC is returned.
        Any filter described in the docs for [DescribeVpcs](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeVpcs.html)
        is valid.

    `vpc_id` take precedence over this.

-   `x509_upload_path` (string) - The path on the remote machine where the X509
    certificate will be uploaded. This path must already exist and be writable.
    X509 certificates are uploaded after provisioning is run, so it is perfectly
    okay to create this directory as part of the provisioning process. Defaults to
    `/tmp`.

-   `windows_password_timeout` (string) - The timeout for waiting for a Windows
    password for Windows instances. Defaults to 20 minutes. Example value: `10m`

## Basic Example

Here is a basic example. It is completely valid except for the access keys:

``` json
{
  "type": "amazon-instance",
  "access_key": "YOUR KEY HERE",
  "secret_key": "YOUR SECRET KEY HERE",
  "region": "us-east-1",
  "source_ami": "ami-d9d6a6b0",
  "instance_type": "m1.small",
  "ssh_username": "ubuntu",

  "account_id": "0123-4567-0890",
  "s3_bucket": "packer-images",
  "x509_cert_path": "x509.cert",
  "x509_key_path": "x509.key",
  "x509_upload_path": "/tmp",

  "ami_name": "packer-quick-start {{timestamp}}"
}
```

-&gt; **Note:** Packer can also read the access key and secret access key from
environmental variables. See the configuration reference in the section above
for more information on what environmental variables Packer will look for.

## Accessing the Instance to Debug

If you need to access the instance to debug for some reason, run this builder
with the `-debug` flag. In debug mode, the Amazon builder will save the private
key in the current directory and will output the DNS or IP information as well.
You can use this information to access the instance as it is running.

## Build template data

The available variables are:

- `BuildRegion` - The region (for example `eu-central-1`) where Packer is building the AMI.
- `SourceAMI` - The source AMI ID (for example `ami-a2412fcd`) used to build the AMI.
- `SourceAMIName` - The source AMI Name (for example `ubuntu/images/ebs-ssd/ubuntu-xenial-16.04-amd64-server-20180306`) used to build the AMI.
- `SourceAMITags` - The source AMI Tags, as a `map[string]string` object.

## Custom Bundle Commands

A lot of the process required for creating an instance-store backed AMI involves
commands being run on the actual source instance. Specifically, the
`ec2-bundle-vol` and `ec2-upload-bundle` commands must be used to bundle the
root filesystem and upload it, respectively.

Each of these commands have a lot of available flags. Instead of exposing each
possible flag as a template configuration option, the instance-store AMI builder
for Packer lets you customize the entire command used to bundle and upload the
AMI.

These are configured with `bundle_vol_command` and `bundle_upload_command`. Both
of these configurations are [configuration
templates](/docs/templates/engine.html) and have support for
their own set of template variables.

### Bundle Volume Command

The default value for `bundle_vol_command` is shown below. It is split across
multiple lines for convenience of reading. The bundle volume command is
responsible for executing `ec2-bundle-vol` in order to store and image of the
root filesystem to use to create the AMI.

``` text
sudo -i -n ec2-bundle-vol \
  -k {{.KeyPath}}  \
  -u {{.AccountId}} \
  -c {{.CertPath}} \
  -r {{.Architecture}} \
  -e {{.PrivatePath}}/* \
  -d {{.Destination}} \
  -p {{.Prefix}} \
  --batch \
  --no-filter
```

The available template variables should be self-explanatory based on the
parameters they're used to satisfy the `ec2-bundle-vol` command.

~&gt; **Warning!** Some versions of ec2-bundle-vol silently ignore all .pem and
.gpg files during the bundling of the AMI, which can cause problems on some
systems, such as Ubuntu. You may want to customize the bundle volume command to
include those files (see the `--no-filter` option of `ec2-bundle-vol`).

### Bundle Upload Command

The default value for `bundle_upload_command` is shown below. It is split across
multiple lines for convenience of reading. Access key and secret key are omitted
if using instance profile. The bundle upload command is responsible for taking
the bundled volume and uploading it to S3.

``` text
sudo -i -n ec2-upload-bundle \
  -b {{.BucketName}} \
  -m {{.ManifestPath}} \
  -a {{.AccessKey}} \
  -s {{.SecretKey}} \
  -d {{.BundleDirectory}} \
  --batch \
  --region {{.Region}} \
  --retry
```

The available template variables should be self-explanatory based on the
parameters they're used to satisfy the `ec2-upload-bundle` command.
Additionally, `{{.Token}}` is available when overriding this command. You must
create your own bundle command with the addition of `-t {{.Token}} ` if you are
assuming a role.

#### Bundle Upload Permissions

The `ec2-upload-bundle` requires a policy document that looks something like this:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:ListBucket",
                "s3:GetBucketLocation",
                "s3:PutObjectAcl"
            ],
            "Resource": "*"
        }
    ]
}
```

You may wish to constrain the resource to a specific bucket.
