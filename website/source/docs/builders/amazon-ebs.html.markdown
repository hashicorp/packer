---
layout: "docs"
---

# Amazon AMI Builder

Type: `amazon-ebs`

The `amazon-ebs` builder is able to create Amazon AMIs backed by EBS
volumes for use in [EC2](http://aws.amazon.com/ec2/). The builder takes
an initial source AMI, runs any provisioning necesary on the instance,
and snapshots it into a reusable AMI.

Amazon supports two types of AMIs: EBS-backed and instance-store. Instance
store AMIs are considerably harder to create, requiring many platform-specific
steps that can often take a very long time. EBS-backed AMIs, on the hand,
only require a source AMI to exist. This builder only builds EBS-backed
instances, because they are easier to create, especially across many
platforms running Packer.

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

Required:

* `access_key` (string) - The access key used to communicate with AWS.
  If not specified, Packer will attempt to read this from environmental
  variables `AWS_ACCESS_KEY_ID` or `AWS_ACCESS_KEY` (in that order).

* `ami_name` (string) - The name of the resulting AMI that will appear
  when managing AMIs in the AWS console or via APIs. This must be unique.
  To help make this unique, certain template parameters are available for
  this value, which are documented below.

* `instance_type` (string) - The EC2 instance type to use while building
  the AMI, such as "m1.small".

* `region` (string) - The name of the region, such as "us-east-1", in which
  to launch the EC2 instance to create the AMI.

* `secret_key` (string) - The secret key used to communicate with AWS.
  If not specified, Packer will attempt to read this from environmental
  variables `AWS_SECRET_ACCESS_KEY` or `AWS_SECRET_KEY` (in that order).

* `source_ami` (string) - The initial AMI used as a base for the newly
  created machine.

* `ssh_username` (string) - The username to use in order to communicate
  over SSH to the running machine.

Optional:

* `ssh_port` (int) - The port that SSH will be available on. This defaults
  to port 22.

* `ssh_timeout` (string) - The time to wait for SSH to become available
  before timing out. The format of this value is a duration such as "5s"
  or "5m". The default SSH timeout is "1m", or one minute.

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
  "ami_name": "packer-quick-start {{.CreateTime}}"
}
</pre>

<div class="alert alert-block alert-info">
<strong>Note:</strong> Packer can also read the access key and secret
access key from environmental variables. See the configuration reference in
the section above for more information on what environmental variables Packer
will look for.
</div>

## AMI Name Variables

The AMI name specified by the `ami_name` configuration variable is actually
treated as a [configuration template](/docs/templates/configuration-templates.html).
Packer provides a set of variables that it will replace
within the AMI name. This helps ensure the AMI name is unique, as AWS requires.

The available variables are shown below:

* `CreateTime` - This will be replaced with the Unix timestamp of when
  the AMI was built.
