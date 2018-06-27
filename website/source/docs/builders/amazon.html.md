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

## Authentication

The AWS provider offers a flexible means of providing credentials for
authentication. The following methods are supported, in this order, and
explained below:

-   Static credentials
-   Environment variables
-   Shared credentials file
-   EC2 Role

### Static Credentials

Static credentials can be provided in the form of an access key id and secret.
These look like:

```json
{
    "access_key": "AKIAIOSFODNN7EXAMPLE",
    "secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    "region": "us-east-1",
    "type": "amazon-ebs"
}
```

### Environment variables

You can provide your credentials via the `AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY`, environment variables, representing your AWS Access
Key and AWS Secret Key, respectively. Note that setting your AWS credentials
using either these environment variables will override the use of
`AWS_SHARED_CREDENTIALS_FILE` and `AWS_PROFILE`. The `AWS_DEFAULT_REGION` and
`AWS_SESSION_TOKEN` environment variables are also used, if applicable:


Usage:

```
$ export AWS_ACCESS_KEY_ID="anaccesskey"
$ export AWS_SECRET_ACCESS_KEY="asecretkey"
$ export AWS_DEFAULT_REGION="us-west-2"
$ packer build packer.json
```

### Shared Credentials file

You can use an AWS credentials file to specify your credentials. The default
location is &#36;HOME/.aws/credentials on Linux and OS X, or
"%USERPROFILE%.aws\credentials" for Windows users. If we fail to detect
credentials inline, or in the environment, Packer will check this location. You
can optionally specify a different location in the configuration by setting the
environment with the `AWS_SHARED_CREDENTIALS_FILE` variable.

The format for the credentials file is like so

```
[default]
aws_access_key_id=<your access key id>
aws_secret_access_key=<your secret access key>
```

You may also configure the profile to use by setting the `profile`
configuration option, or setting the `AWS_PROFILE` environment variable:

```json
{
    "profile": "customprofile",
    "region": "us-east-1",
    "type": "amazon-ebs"
}
```


### IAM Task or Instance Role

Finally, Packer will use credentials provided by the task's or instance's IAM
role, if it has one.

This is a preferred approach over any other when running in EC2 as you can
avoid hard coding credentials. Instead these are leased on-the-fly by Packer,
which reduces the chance of leakage.

The following policy document provides the minimal set permissions necessary
for Packer to work:

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
        "ec2:DeleteKeyPair",
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

Note that if you'd like to create a spot instance, you must also add:

``` json
ec2:RequestSpotInstances,
ec2:CancelSpotInstanceRequests,
ec2:DescribeSpotInstanceRequests
```

If you have the `spot_price` parameter set to `auto`, you must also add:

``` json
ec2:DescribeSpotPriceHistory
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
