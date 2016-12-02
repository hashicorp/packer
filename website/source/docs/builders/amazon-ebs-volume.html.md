---
description: |
    The `amazon-ebsvolume` Packer builder is like the EBS builder, but is
    intended to create EBS volumes rather than a machine image.
layout: docs
page_title: 'Amazon EBS Volume Builder'
...

# EBS Volume Builder

Type: `amazon-ebsvolume`

The `amazon-ebsvolume` Packer builder is able to create Amazon Elastic Block
Store volumes which are prepopulated with filesystems or data. 

This builder builds EBS volumes by launching an EC2 instance from a source AMI,
provisioning that running machine, and then destroying the source machine, keeping
the volumes intact.

This is all done in your own AWS account. The builder will create temporary
keypairs, security group rules, etc. that provide it temporary access to the
instance while the image is being created. 

The builder does *not* manage EBS Volumes. Once it creates volumes and stores
it in your account, it is up to you to use, delete, etc. the volumes.

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

-   `instance_type` (string) - The EC2 instance type to use while building the
    AMI, such as "m1.small".

-   `region` (string) - The name of the region, such as "us-east-1", in which to
    launch the EC2 instance to create the AMI.

-   `secret_key` (string) - The secret key used to communicate with AWS. [Learn
    how to set this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `source_ami` (string) - The initial AMI used as a base for the newly
    created machine. `source_ami_filter` may be used instead to populate this
    automatically.

### Optional:

-   `ebs_volumes` (array of block device mappings) - Add the block
    device mappings to the AMI. The block device mappings allow for keys:

    -   `device_name` (string) - The device name exposed to the instance (for
         example, "/dev/sdh" or "xvdh"). Required when specifying `volume_size`.
    -   `delete_on_termination` (boolean) - Indicates whether the EBS volume is
        deleted on instance termination
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
    -   `volume_type` (string) - The volume type. gp2 for General Purpose (SSD)
        volumes, io1 for Provisioned IOPS (SSD) volumes, and standard for Magnetic
        volumes
    -   `tags` (map) - Tags to apply to the volume. These are retained after the
        builder completes.

-   `associate_public_ip_address` (boolean) - If using a non-default VPC, public
    IP addresses are not provided by default. If this is toggled, your new
    instance will get a Public IP.

-   `availability_zone` (string) - Destination availability zone to launch
    instance in. Leave this empty to allow Amazon to auto-assign.

-   `ebs_optimized` (boolean) - Mark instance as [EBS
    Optimized](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSOptimized.html).
    Default `false`.

-   `enhanced_networking` (boolean) - Enable enhanced
    networking (SriovNetSupport) on HVM-compatible AMIs. If true, add
    `ec2:ModifyInstanceAttribute` to your AWS IAM policy.

-   `iam_instance_profile` (string) - The name of an [IAM instance
    profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/instance-profiles.html)
    to launch the EC2 instance with.

-   `run_tags` (object of key/value strings) - Tags to apply to the instance
    that is *launched* to create the AMI. These tags are *not* applied to the
    resulting AMI unless they're duplicated in `tags`.

-   `security_group_id` (string) - The ID (*not* the name) of the security group
    to assign to the instance. By default this is not set and Packer will
    automatically create a new temporary security group to allow SSH access.
    Note that if this is specified, you must be sure the security group allows
    access to the `ssh_port` given below.

-   `security_group_ids` (array of strings) - A list of security groups as
    described above. Note that if this is specified, you must omit the
    `security_group_id`.

-   `shutdown_behaviour` (string) - Automatically terminate instances on shutdown
    incase packer exits ungracefully. Possible values are "stop" and "terminate",
    default is stop.

-   `skip_region_validation` (boolean) - Set to true if you want to skip 
    validation of the region configuration option.  Defaults to false.

-   `snapshot_groups` (array of strings) - A list of groups that have access to
    create volumes from the snapshot(s). By default no groups have permission to create
    volumes form the snapshot(s). `all` will make the snapshot publicly accessible.

-   `snapshot_users` (array of strings) - A list of account IDs that have access to
    create volumes from the snapshot(s). By default no additional users other than the
    user creating the AMI has permissions to create volumes from the backing snapshot(s).

-   `source_ami_filter` (object) - Filters used to populate the `source_ami` field.
    Example:

    ``` {.javascript}
    "source_ami_filter": {
        "filters": {
          "virtualization-type": "hvm",
          "name": "*ubuntu-xenial-16.04-amd64-server-*",
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

    -   `owners` (array of strings) - This scopes the AMIs to certain Amazon account IDs.
         This is helpful to limit the AMIs to a trusted third party, or to your own account.

    -   `most_recent` (bool) - Selects the newest created image when true.
         This is most useful for selecting a daily distro build.

-   `spot_price` (string) - The maximum hourly price to pay for a spot instance
    to create the AMI. Spot instances are a type of instance that EC2 starts
    when the current spot price is less than the maximum price you specify. Spot
    price will be updated based on available spot instance capacity and current
    spot instance requests. It may save you some costs. You can set this to
    "auto" for Packer to automatically discover the best spot price or to "0"
    to use an on demand instance (default).

-   `spot_price_auto_product` (string) - Required if `spot_price` is set
    to "auto". This tells Packer what sort of AMI you're launching to find the
    best spot price. This must be one of: `Linux/UNIX`, `SUSE Linux`, `Windows`,
    `Linux/UNIX (Amazon VPC)`, `SUSE Linux (Amazon VPC)`, `Windows (Amazon VPC)`

-   `ssh_keypair_name` (string) - If specified, this is the key that will be
    used for SSH with the machine. By default, this is blank, and Packer will
    generate a temporary keypair unless
    [`ssh_password`](/docs/templates/communicator.html#ssh_password) is used.
    [`ssh_private_key_file`](/docs/templates/communicator.html#ssh_private_key_file)
    must be specified with this.

-   `ssh_private_ip` (boolean) - If true, then SSH will always use the private
    IP if available.

-   `subnet_id` (string) - If using VPC, the ID of the subnet, such as
    "subnet-12345def", where Packer will launch the EC2 instance. This field is
    required if you are using an non-default VPC.

-   `temporary_key_pair_name` (string) - The name of the temporary keypair
    to generate. By default, Packer generates a name with a UUID.

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
    to be set.

-   `windows_password_timeout` (string) - The timeout for waiting for a Windows
    password for Windows instances. Defaults to 20 minutes. Example value: "10m"

## Basic Example

```
{
   "type" : "amazon-ebsvolume",
   "secret_key" : "YOUR SECRET KEY HERE",
   "access_key" : "YOUR KEY HERE",
   "region" : "us-east-1",
   "ssh_username" : "ubuntu",
   "instance_type" : "t2.medium",
   "source_ami" : "ami-40d28157",
   "ebs_volumes" : [
      {
         "volume_type" : "gp2",
         "device_name" : "/dev/xvdf",
         "delete_on_termination" : false,
         "tags" : {
            "zpool" : "data",
            "Name" : "Data1"
         },
         "volume_size" : 10
      },
      {
         "volume_type" : "gp2",
         "device_name" : "/dev/xvdg",
         "tags" : {
            "zpool" : "data",
            "Name" : "Data2"
         },
         "delete_on_termination" : false,
         "volume_size" : 10
      },
      {
         "volume_size" : 10,
         "tags" : {
            "Name" : "Data3",
            "zpool" : "data"
         },
         "delete_on_termination" : false,
         "device_name" : "/dev/xvdh",
         "volume_type" : "gp2"
      }
   ]
}
```

-&gt; **Note:** Packer can also read the access key and secret access key from
environmental variables. See the configuration reference in the section above
for more information on what environmental variables Packer will look for.

Further information on locating AMI IDs and their relationship to instance 
types and regions can be found in the AWS EC2 Documentation 
[for Linux](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/finding-an-ami.html) 
or [for Windows](http://docs.aws.amazon.com/AWSEC2/latest/WindowsGuide/finding-an-ami.html).

## Accessing the Instance to Debug

If you need to access the instance to debug for some reason, run the builder
with the `-debug` flag. In debug mode, the Amazon builder will save the private
key in the current directory and will output the DNS or IP information as well.
You can use this information to access the instance as it is running.

-&gt; **Note:** Packer uses pre-built AMIs as the source for building images.
These source AMIs may include volumes that are not flagged to be destroyed on
termination of the instance building the new image. In addition to those volumes
created by this builder, any volumes inn the source AMI which are not marked for
deletion on termination will remain in your account.
