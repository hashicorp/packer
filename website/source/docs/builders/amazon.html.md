---
description: |
    Packer is able to create Amazon AMIs. To achieve this, Packer comes with
    multiple builders depending on the strategy you want to use to build the AMI.
layout: docs
page_title: 'Amazon AMI - Builders'
sidebar_current: 'docs-builders-amazon'
---

# Amazon AMI Builder

Packer is able to create Amazon AMIs. To achieve this, Packer comes with
multiple builders depending on the strategy you want to use to build the AMI.
Packer supports the following builders at the moment:

-   [amazon-ebs](/docs/builders/amazon-ebs.html) - Create EBS-backed AMIs by
    launching a source AMI and re-packaging it into a new AMI
    after provisioning. If in doubt, use this builder, which is the easiest to
    get started with.

-   [amazon-instance](/docs/builders/amazon-instance.html) - Create
    instance-store AMIs by launching and provisioning a source instance, then
    rebundling it and uploading it to S3.

-   [amazon-chroot](/docs/builders/amazon-chroot.html) - Create EBS-backed AMIs
    from an existing EC2 instance by mounting the root device and using a
    [Chroot](https://en.wikipedia.org/wiki/Chroot) environment to provision
    that device. This is an **advanced builder and should not be used by
    newcomers**. However, it is also the fastest way to build an EBS-backed AMI
    since no new EC2 instance needs to be launched.

-   [amazon-ebssurrogate](/docs/builders/amazon-ebssurrogate.html) - Create EBS
    -backed AMIs from scratch. Works similarly to the `chroot` builder but does
    not require running in AWS. This is an **advanced builder and should not be
    used by newcomers**.

-&gt; **Don't know which builder to use?** If in doubt, use the [amazon-ebs
builder](/docs/builders/amazon-ebs.html). It is much easier to use and Amazon
generally recommends EBS-backed images nowadays.

# Amazon EBS Volume Builder

Packer is able to create Amazon EBS Volumes which are preinitialized with a
filesystem and data.

-   [amazon-ebsvolume](/docs/builders/amazon-ebsvolume.html) - Create EBS volumes
    by launching a source AMI with block devices mapped. Provision the instance,
    then destroy it, retaining the EBS volumes.

<span id="specifying-amazon-credentials"></span>

## Specifying Amazon Credentials

When you use any of the amazon builders, you must provide credentials to the API
in the form of an access key id and secret. These look like:

    access key id:     AKIAIOSFODNN7EXAMPLE
    secret access key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

If you use other AWS tools you may already have these configured. If so, packer
will try to use them, *unless* they are specified in your packer template.
Credentials are resolved in the following order:

1.  Values hard-coded in the packer template are always authoritative.
2.  *Variables* in the packer template may be resolved from command-line flags
    or from environment variables. Please read about [User
    Variables](https://www.packer.io/docs/templates/user-variables.html)
    for details.
3.  If no credentials are found, packer falls back to automatic lookup.

### Automatic Lookup

Packer depends on the [AWS
SDK](https://aws.amazon.com/documentation/sdk-for-go/) to perform automatic
lookup using *credential chains*. In short, the SDK looks for credentials in
the following order:

1.  Environment variables.
2.  Shared credentials file.
3.  If your application is running on an Amazon EC2 instance, IAM role for Amazon EC2.

Please refer to the SDK's documentation on [specifying
credentials](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html#specifying-credentials)
for more information.

## Using an IAM Task or Instance Role

If AWS keys are not specified in the template, a
[shared credentials file](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files)
or through environment variables Packer will use credentials provided by
the task's or instance's IAM role, if it has one.

The following policy document provides the minimal set permissions necessary for
Packer to work:

``` json
{
  "Version": "2012-10-17",
  "Statement": [{
      "Effect": "Allow",
      "Action" : [
        "ec2:AttachVolume",
        "ec2:AuthorizeSecurityGroupIngress",
        "ec2:CopyImage",
        "ec2:CreateImage",
        "ec2:CreateKeypair",
        "ec2:CreateSecurityGroup",
        "ec2:CreateSnapshot",
        "ec2:CreateTags",
        "ec2:CreateVolume",
        "ec2:DeleteKeypair",
        "ec2:DeleteSecurityGroup",
        "ec2:DeleteSnapshot",
        "ec2:DeleteVolume",
        "ec2:DeregisterImage",
        "ec2:DescribeImageAttribute",
        "ec2:DescribeImages",
        "ec2:DescribeInstances",
        "ec2:DescribeRegions",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSnapshots",
        "ec2:DescribeSubnets",
        "ec2:DescribeTags",
        "ec2:DescribeVolumes",
        "ec2:DetachVolume",
        "ec2:GetPasswordData",
        "ec2:ModifyImageAttribute",
        "ec2:ModifyInstanceAttribute",
        "ec2:ModifySnapshotAttribute",
        "ec2:RegisterImage",
        "ec2:RunInstances",
        "ec2:StopInstances",
        "ec2:TerminateInstances"
      ],
      "Resource" : "*"
  }]
}
```

## Troubleshooting

### Attaching IAM Policies to Roles

IAM policies can be associated with users or roles. If you use packer with IAM
roles, you may encounter an error like this one:

    ==> amazon-ebs: Error launching source instance: You are not authorized to perform this operation.

You can read more about why this happens on the [Amazon Security
Blog](https://blogs.aws.amazon.com/security/post/Tx3M0IFB5XBOCQX/Granting-Permission-to-Launch-EC2-Instances-with-IAM-Roles-PassRole-Permission).
The example policy below may help packer work with IAM roles. Note that this
example provides more than the minimal set of permissions needed for packer to
work, but specifics will depend on your use-case.

``` json
{
    "Sid": "PackerIAMPassRole",
    "Effect": "Allow",
    "Action": "iam:PassRole",
    "Resource": [
        "*"
    ]
}
```

### Checking that system time is current

Amazon uses the current time as part of the [request signing
process](http://docs.aws.amazon.com/general/latest/gr/sigv4_signing.html). If
your system clock is too skewed from the current time, your requests might
fail. If that's the case, you might see an error like this:

    ==> amazon-ebs: Error querying AMI: AuthFailure: AWS was not able to validate the provided access credentials

If you suspect your system's date is wrong, you can compare it against
<http://www.time.gov/>. On Linux/OS X, you can run the `date` command to get the
current time. If you're on Linux, you can try setting the time with ntp by
running `sudo ntpd -q`.
