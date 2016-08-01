---
description: |
    The `amazon-instance` Packer builder is able to create Amazon AMIs backed by
    instance storage as the root device. For more information on the difference
    between instance storage and EBS-backed instances, see the storage for the root
    device section in the EC2 documentation.
layout: docs
page_title: 'Amazon AMI Builder (instance-store)'
...

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
AWS account. The builder will create temporary keypairs, security group rules,
etc. that provide it temporary access to the instance while the image is being
created. This simplifies configuration quite a bit.

The builder does *not* manage AMIs. Once it creates an AMI and stores it in
your account, it is up to you to use, delete, etc. the AMI.

-> **Note:** This builder requires that the [Amazon EC2 AMI
Tools](https://aws.amazon.com/developertools/368) are installed onto the
machine. This can be done within a provisioner, but must be done before the
builder finishes running.

~> Instance builds are not supported for Windows. Use [`amazon-ebs`](amazon-ebs.html) instead.

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

-   `account_id` (string) - Your AWS account ID. This is required for bundling
    the AMI. This is *not the same* as the access key. You can find your account
    ID in the security credentials page of your AWS account.

-   `ami_name` (string) - The name of the resulting AMI that will appear when
    managing AMIs in the AWS console or via APIs. This must be unique. To help
    make this unique, use a function like `timestamp` (see [configuration
    templates](/docs/templates/configuration-templates.html) for more info)

-   `instance_type` (string) - The EC2 instance type to use while building the
    AMI, such as "m1.small".

-   `region` (string) - The name of the region, such as "us-east-1", in which to
    launch the EC2 instance to create the AMI.

-   `s3_bucket` (string) - The name of the S3 bucket to upload the AMI. This
    bucket will be created if it doesn't exist.

-   `secret_key` (string) - The secret key used to communicate with AWS. [Learn
    how to set this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `source_ami` (string) - The initial AMI used as a base for the newly
    created machine.

-   `ssh_username` (string) - The username to use in order to communicate over
    SSH to the running machine.

-   `x509_cert_path` (string) - The local path to a valid X509 certificate for
    your AWS account. This is used for bundling the AMI. This X509 certificate
    must be registered with your account from the security credentials page in
    the AWS console.

-   `x509_key_path` (string) - The local path to the private key for the X509
    certificate specified by `x509_cert_path`. This is used for bundling
    the AMI.

### Optional:

-   `ami_block_device_mappings` (array of block device mappings) - Add the block
    device mappings to the AMI. The block device mappings allow for keys:
    -   `delete_on_termination` (boolean) - Indicates whether the EBS volume is
        deleted on instance termination
    -   `device_name` (string) - The device name exposed to the instance (for
        example, "/dev/sdh" or "xvdh"). Required when specifying `volume_size`.
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
-   `ami_description` (string) - The description to set for the
    resulting AMI(s). By default this description is empty.

-   `ami_groups` (array of strings) - A list of groups that have access to
    launch the resulting AMI(s). By default no groups have permission to launch
    the AMI. `all` will make the AMI publicly accessible. AWS currently doesn't
    accept any value other than "all".

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
    "paravirtual" (default) or "hvm".

-   `associate_public_ip_address` (boolean) - If using a non-default VPC, public
    IP addresses are not provided by default. If this is toggled, your new
    instance will get a Public IP.

-   `availability_zone` (string) - Destination availability zone to launch
    instance in. Leave this empty to allow Amazon to auto-assign.

-   `bundle_destination` (string) - The directory on the running instance where
    the bundled AMI will be saved prior to uploading. By default this is "/tmp".
    This directory must exist and be writable.

-   `bundle_prefix` (string) - The prefix for files created from bundling the
    root volume. By default this is "image-{{timestamp}}". The `timestamp`
    variable should be used to make sure this is unique, otherwise it can
    collide with other created AMIs by Packer in your account.

-   `bundle_upload_command` (string) - The command to use to upload the
    bundled volume. See the "custom bundle commands" section below for
    more information.

-   `bundle_vol_command` (string) - The command to use to bundle the volume. See
    the "custom bundle commands" section below for more information.

-   `ebs_optimized` (boolean) - Mark instance as [EBS
    Optimized](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSOptimized.html).
    Default `false`.

-   `enhanced_networking` (boolean) - Enable enhanced
    networking (SriovNetSupport) on HVM-compatible AMIs. If true, add
    `ec2:ModifyInstanceAttribute` to your AWS IAM policy.

-   `force_deregister` (boolean) - Force Packer to first deregister an existing
    AMI if one with the same name already exists. Default `false`.

-   `iam_instance_profile` (string) - The name of an [IAM instance
    profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/instance-profiles.html)
    to launch the EC2 instance with.

-   `launch_block_device_mappings` (array of block device mappings) - Add the
    block device mappings to the launch instance. The block device mappings are
    the same as `ami_block_device_mappings` above.

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

-   `skip_region_validation` (boolean) - Set to true if you want to skip 
    validation of the region configuration option.  Defaults to false.

-   `spot_price` (string) - The maximum hourly price to launch a spot instance
    to create the AMI. It is a type of instances that EC2 starts when the
    maximum price that you specify exceeds the current spot price. Spot price
    will be updated based on available spot instance capacity and current spot
    Instance requests. It may save you some costs. You can set this to "auto"
    for Packer to automatically discover the best spot price or to "0" to use
    an on demand instance (default).

-   `spot_price_auto_product` (string) - Required if `spot_price` is set
    to "auto". This tells Packer what sort of AMI you're launching to find the
    best spot price. This must be one of: `Linux/UNIX`, `SUSE Linux`, `Windows`,
    `Linux/UNIX (Amazon VPC)`, `SUSE Linux (Amazon VPC)`, `Windows (Amazon VPC)`

-   `ssh_keypair_name` (string) - If specified, this is the key that will be
    used for SSH with the machine. The key must match a key pair name loaded
    up into Amazon EC2.  By default, this is blank, and Packer will
    generate a temporary keypair.
    [`ssh_private_key_file`](/docs/templates/communicator.html#ssh_private_key_file)
    must be specified when `ssh_keypair_name` is utilized.

-   `ssh_private_ip` (boolean) - If true, then SSH will always use the private
    IP if available.

-   `subnet_id` (string) - If using VPC, the ID of the subnet, such as
    "subnet-12345def", where Packer will launch the EC2 instance. This field is
    required if you are using an non-default VPC.

-   `tags` (object of key/value strings) - Tags applied to the AMI.

-   `temporary_key_pair_name` (string) - The name of the temporary keypair
    to generate. By default, Packer generates a name with a UUID.

-   `user_data` (string) - User data to apply when launching the instance. Note
    that you need to be careful about escaping characters due to the templates
    being JSON. It is often more convenient to use `user_data_file`, instead.

-   `user_data_file` (string) - Path to a file that will be used for the user
    data when launching the instance.

-   `vpc_id` (string) - If launching into a VPC subnet, Packer needs the VPC ID
    in order to create a temporary security group within the VPC.

-   `x509_upload_path` (string) - The path on the remote machine where the X509
    certificate will be uploaded. This path must already exist and be writable.
    X509 certificates are uploaded after provisioning is run, so it is perfectly
    okay to create this directory as part of the provisioning process.

-   `windows_password_timeout` (string) - The timeout for waiting for a Windows
    password for Windows instances. Defaults to 20 minutes. Example value: "10m"

## Basic Example

Here is a basic example. It is completely valid except for the access keys:

``` {.javascript}
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

If you need to access the instance to debug for some reason, run the builder
with the `-debug` flag. In debug mode, the Amazon builder will save the private
key in the current directory and will output the DNS or IP information as well.
You can use this information to access the instance as it is running.

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
templates](/docs/templates/configuration-templates.html) and have support for
their own set of template variables.

### Bundle Volume Command

The default value for `bundle_vol_command` is shown below. It is split across
multiple lines for convenience of reading. The bundle volume command is
responsible for executing `ec2-bundle-vol` in order to store and image of the
root filesystem to use to create the AMI.

``` {.text}
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

\~&gt; **Warning!** Some versions of ec2-bundle-vol silently ignore all .pem and
.gpg files during the bundling of the AMI, which can cause problems on some
systems, such as Ubuntu. You may want to customize the bundle volume command to
include those files (see the `--no-filter` option of ec2-bundle-vol).

### Bundle Upload Command

The default value for `bundle_upload_command` is shown below. It is split across
multiple lines for convenience of reading. The bundle upload command is
responsible for taking the bundled volume and uploading it to S3.

``` {.text}
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
