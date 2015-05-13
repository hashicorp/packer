---
layout: "docs"
page_title: "Amazon AMI Builder"
description: |-
  Packer is able to create Amazon AMIs. To achieve this, Packer comes with multiple builders depending on the strategy you want to use to build the AMI.
---

# Amazon AMI Builder

Packer is able to create Amazon AMIs. To achieve this, Packer comes with
multiple builders depending on the strategy you want to use to build the
AMI. Packer supports the following builders at the moment:

* [amazon-ebs](/docs/builders/amazon-ebs.html) - Create EBS-backed AMIs
  by launching a source AMI and re-packaging it into a new AMI after
  provisioning. If in doubt, use this builder, which is the easiest to get
  started with.

* [amazon-instance](/docs/builders/amazon-instance.html) - Create
  instance-store AMIs by launching and provisioning a source instance, then
  rebundling it and uploading it to S3.

* [amazon-chroot](/docs/builders/amazon-chroot.html) - Create EBS-backed AMIs
  from an existing EC2 instance by mounting the root device and using a
  [Chroot](http://en.wikipedia.org/wiki/Chroot) environment to provision
  that device. This is an **advanced builder and should not be used by
  newcomers**. However, it is also the fastest way to build an EBS-backed
  AMI since no new EC2 instance needs to be launched.

-> **Don't know which builder to use?** If in doubt, use the
[amazon-ebs builder](/docs/builders/amazon-ebs.html). It is
much easier to use and Amazon generally recommends EBS-backed images nowadays.

## Using an IAM Instance Profile

If AWS keys are not specified in the template, a [credentials](http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files) file or through environment variables
Packer will use credentials provided by the instance's IAM profile, if it has one.

The following policy document provides the minimal set permissions necessary for Packer to work:

```javascript
{
  "Statement": [{
      "Effect": "Allow",
      "Action" : [
        "ec2:AttachVolume",
        "ec2:CreateVolume",
        "ec2:DeleteVolume",
        "ec2:CreateKeypair",
        "ec2:DeleteKeypair",
        "ec2:CreateSecurityGroup",
        "ec2:DeleteSecurityGroup",
        "ec2:AuthorizeSecurityGroupIngress",
        "ec2:CreateImage",
        "ec2:RunInstances",
        "ec2:TerminateInstances",
        "ec2:StopInstances",
        "ec2:DescribeVolumes",
        "ec2:DetachVolume",
        "ec2:DescribeInstances",
        "ec2:CreateSnapshot",
        "ec2:DeleteSnapshot",
        "ec2:DescribeSnapshots",
        "ec2:DescribeImages",
        "ec2:RegisterImage",
        "ec2:CreateTags",
        "ec2:ModifyImageAttribute"
      ],
      "Resource" : "*"
  }]
}
```
