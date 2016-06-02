---
description: |
    Packer is able to create Amazon AMIs. To achieve this, Packer comes with
    multiple builders depending on the strategy you want to use to build the AMI.
layout: docs
page_title: Amazon AMI Builder
...

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

-&gt; **Don't know which builder to use?** If in doubt, use the [amazon-ebs
builder](/docs/builders/amazon-ebs.html). It is much easier to use and Amazon
generally recommends EBS-backed images nowadays.

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

If no AWS credentials are found in a packer template, we proceed on to the
following steps:

1.  Lookup via environment variables.
    -   First `AWS_ACCESS_KEY_ID`, then `AWS_ACCESS_KEY`
    -   First `AWS_SECRET_ACCESS_KEY`, then `AWS_SECRET_KEY`

2.  Look for [local AWS configuration
    files](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files)
    -   First `~/.aws/credentials`
    -   Next based on `AWS_PROFILE`

3.  Lookup an IAM role for the current EC2 instance (if you're running in EC2)

\~&gt; **Subtle details of automatic lookup may change over time.** The most
reliable way to specify your configuration is by setting them in template
variables (directly or indirectly), or by using the `AWS_ACCESS_KEY_ID` and
`AWS_SECRET_ACCESS_KEY` environment variables.

Environment variables provide the best portability, allowing you to run your
packer build on your workstation, in Atlas, or on another build server.

## Using an IAM Instance Profile

If AWS keys are not specified in the template, a
[credentials](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-config-files)
file or through environment variables Packer will use credentials provided by
the instance's IAM profile, if it has one.

The following policy document provides the minimal set permissions necessary for
Packer to work:

``` {.javascript}
{
  "Statement": [{
      "Effect": "Allow",
      "Action" : [
        "ec2:AttachVolume",
        "ec2:CreateVolume",
        "ec2:DeleteVolume",
        "ec2:CreateKeypair",
        "ec2:DeleteKeypair",
        "ec2:DescribeSubnets",
        "ec2:CreateSecurityGroup",
        "ec2:DeleteSecurityGroup",
        "ec2:AuthorizeSecurityGroupIngress",
        "ec2:CreateImage",
        "ec2:CopyImage",
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
        "ec2:ModifyImageAttribute",
        "ec2:GetPasswordData",
        "ec2:DescribeTags",
        "ec2:DescribeImageAttribute",
        "ec2:CopyImage",
        "ec2:DescribeRegions"
      ],
      "Resource" : "*"
  }]
}
```

## Troubleshooting

### Attaching IAM Policies to Roles

IAM policies can be associated with user or roles. If you use packer with IAM
roles, you may encounter an error like this one:

    ==> amazon-ebs: Error launching source instance: You are not authorized to perform this operation.

You can read more about why this happens on the [Amazon Security
Blog](https://blogs.aws.amazon.com/security/post/Tx3M0IFB5XBOCQX/Granting-Permission-to-Launch-EC2-Instances-with-IAM-Roles-PassRole-Permission).
The example policy below may help packer work with IAM roles. Note that this
example provides more than the minimal set of permissions needed for packer to
work, but specifics will depend on your use-case.

``` {.json}
{
    "Sid": "PackerIAMPassRole",
    "Effect": "Allow",
    "Action": "iam:PassRole",
    "Resource": [
        "*"
    ]
}
```
