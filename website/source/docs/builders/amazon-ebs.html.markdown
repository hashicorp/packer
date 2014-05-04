---
layout: "docs"
page_title: "Amazon AMI Builder (EBS backed)"
---

# AMI Builder (EBS backed)

Type: `amazon-ebs`

The `amazon-ebs` builder is able to create Amazon AMIs backed by EBS
volumes for use in [EC2](http://aws.amazon.com/ec2/). For more information
on the difference betwen EBS-backed instances and instance-store backed
instances, see the
["storage for the root device" section in the EC2 documentation](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ComponentsAMIs.html#storage-for-the-root-device).

This builder builds an AMI by launching an EC2 instance from a source AMI,
provisioning that running machine, and then creating an AMI from that machine.
This is all done in your own AWS account. The builder will create temporary
keypairs, security group rules, etc. that provide it temporary access to
the instance while the image is being created. This simplifies configuration
quite a bit.

The builder does _not_ manage AMIs. Once it creates an AMI and stores it
in your account, it is up to you to use, delete, etc. the AMI.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

### Required:

* `access_key` (string) - The access key used to communicate with AWS.
  If not specified, Packer will use the environment variables
  `AWS_ACCESS_KEY_ID` or `AWS_ACCESS_KEY` (in that order), if set.

* `ami_name` (string) - The name of the resulting AMI that will appear
  when managing AMIs in the AWS console or via APIs. This must be unique.
  To help make this unique, use a function like `timestamp` (see
  [configuration templates](/docs/templates/configuration-templates.html) for more info)

* `instance_type` (string) - The EC2 instance type to use while building
  the AMI, such as "m1.small".

* `region` (string) - The name of the region, such as "us-east-1", in which
  to launch the EC2 instance to create the AMI.

* `secret_key` (string) - The secret key used to communicate with AWS.
  If not specified, Packer will use the environment variables
  `AWS_SECRET_ACCESS_KEY` or `AWS_SECRET_KEY` (in that order), if set.

* `source_ami` (string) - The initial AMI used as a base for the newly
  created machine.

* `ssh_username` (string) - The username to use in order to communicate
  over SSH to the running machine.

### Optional:

* `ami_block_device_mappings` (array of block device mappings) - Add the block
  device mappings to the AMI. The block device mappings allow for keys:
  "device\_name" (string), "virtual\_name" (string), "snapshot\_id" (string),
  "volume\_type" (string), "volume\_size" (integer), "delete\_on\_termination"
  (boolean), "no\_device" (boolean), and "iops" (integer).

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

* `associate_public_ip_address` (boolean) - If using a non-default VPC, public
  IP addresses are not provided by default. If this is toggled, your new
  instance will get a Public IP.

* `availability_zone` (string) - Destination availability zone to launch instance in.
  Leave this empty to allow Amazon to auto-assign.

* `iam_instance_profile` (string) - The name of an
  [IAM instance profile](http://docs.aws.amazon.com/IAM/latest/UserGuide/instance-profiles.html)
  to launch the EC2 instance with.

* `launch_block_device_mappings` (array of block device mappings) - Add the
  block device mappings to the launch instance. The block device mappings are
  the same as `ami_block_device_mappings` above.

* `run_tags` (object of key/value strings) - Tags to apply to the instance
  that is _launched_ to create the AMI. These tags are _not_ applied to
  the resulting AMI unless they're duplicated in `tags`.

* `security_group_id` (string) - The ID (_not_ the name) of the security
  group to assign to the instance. By default this is not set and Packer
  will automatically create a new temporary security group to allow SSH
  access. Note that if this is specified, you must be sure the security
  group allows access to the `ssh_port` given below.

* `security_group_ids` (array of strings) - A list of security groups as
  described above. Note that if this is specified, you must omit the
  security_group_id.

* `ssh_port` (integer) - The port that SSH will be available on. This defaults
  to port 22.

* `ssh_private_key_file` (string) - Use this ssh private key file instead of
  a generated ssh key pair for connecting to the instance.

* `ssh_timeout` (string) - The time to wait for SSH to become available
  before timing out. The format of this value is a duration such as "5s"
  or "5m". The default SSH timeout is "5m", or five minutes.

* `subnet_id` (string) - If using VPC, the ID of the subnet, such as
  "subnet-12345def", where Packer will launch the EC2 instance.

* `tags` (object of key/value strings) - Tags applied to the AMI.

* `temporary_key_pair_name` (string) - The name of the temporary keypair
  to generate. By default, Packer generates a name with a UUID.

* `user_data` (string) - User data to apply when launching the instance.
  Note that you need to be careful about escaping characters due to the
  templates being JSON. It is often more convenient to use `user_data_file`,
  instead.

* `user_data_file` (string) - Path to a file that will be used for the
  user data when launching the instance.

* `vpc_id` (string) - If launching into a VPC subnet, Packer needs the
  VPC ID in order to create a temporary security group within the VPC.

## Basic Example

Here is a basic example. It is completely valid except for the access keys:

<pre class="prettyprint">
{
  "type": "amazon-ebs",
  "access_key": "YOUR KEY HERE",
  "secret_key": "YOUR SECRET KEY HERE",
  "region": "us-east-1",
  "source_ami": "ami-de0d9eb7",
  "instance_type": "t1.micro",
  "ssh_username": "ubuntu",
  "ami_name": "packer-quick-start {{timestamp}}"
}
</pre>

<div class="alert alert-block alert-info">
<strong>Note:</strong> Packer can also read the access key and secret
access key from environmental variables. See the configuration reference in
the section above for more information on what environmental variables Packer
will look for.
</div>

## Accessing the Instance to Debug

If you need to access the instance to debug for some reason, run the builder
with the `-debug` flag. In debug mode, the Amazon builder will save the
private key in the current directory and will output the DNS or IP information
as well. You can use this information to access the instance as it is
running.

## AMI Block Device Mappings Example

Here is an example using the optional AMI block device mappings. This will add
the /dev/sdb and /dev/sdc block device mappings to the finished AMI.

<pre class="prettyprint">
{
  "type": "amazon-ebs",
  "access_key": "YOUR KEY HERE",
  "secret_key": "YOUR SECRET KEY HERE",
  "region": "us-east-1",
  "source_ami": "ami-de0d9eb7",
  "instance_type": "t1.micro",
  "ssh_username": "ubuntu",
  "ami_name": "packer-quick-start {{timestamp}}",
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
</pre>

## Tag Example

Here is an example using the optional AMI tags. This will add the tags
"OS_Version" and "Release" to the finished AMI.

<pre class="prettyprint">
{
  "type": "amazon-ebs",
  "access_key": "YOUR KEY HERE",
  "secret_key": "YOUR SECRET KEY HERE",
  "region": "us-east-1",
  "source_ami": "ami-de0d9eb7",
  "instance_type": "t1.micro",
  "ssh_username": "ubuntu",
  "ami_name": "packer-quick-start {{timestamp}}",
  "tags": {
    "OS_Version": "Ubuntu",
    "Release": "Latest"
  }
}
</pre>
