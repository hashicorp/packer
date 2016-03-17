---
description: |
    With Packer installed, let's just dive right into it and build our first image.
    Our first image will be an Amazon EC2 AMI with Redis pre-installed. This is just
    an example. Packer can create images for many platforms with anything
    pre-installed.
layout: intro
next_title: Provision
next_url: '/intro/getting-started/provision.html'
page_title: Build an Image
prev_url: '/intro/getting-started/setup.html'
...

# Build an Image

With Packer installed, let's just dive right into it and build our first image.
Our first image will be an [Amazon EC2 AMI](https://aws.amazon.com/ec2/) with
Redis pre-installed. This is just an example. Packer can create images for [many
platforms](/intro/platforms.html) with anything pre-installed.

If you don't have an AWS account, [create one now](https://aws.amazon.com/free/).
For the example, we'll use a "t2.micro" instance to build our image, which
qualifies under the AWS [free-tier](https://aws.amazon.com/free/), meaning it
will be free. If you already have an AWS account, you may be charged some amount
of money, but it shouldn't be more than a few cents.

-&gt; **Note:** If you're not using an account that qualifies under the AWS
free-tier, you may be charged to run these examples. The charge should only be a
few cents, but we're not responsible if it ends up being more.

Packer can build images for [many platforms](/intro/platforms.html) other than
AWS, but AWS requires no additional software installed on your computer and
their [free-tier](https://aws.amazon.com/free/) makes it free to use for most
people. This is why we chose to use AWS for the example. If you're uncomfortable
setting up an AWS account, feel free to follow along as the basic principles
apply to the other platforms as well.

## The Template

The configuration file used to define what image we want built and how is called
a *template* in Packer terminology. The format of a template is simple
[JSON](http://www.json.org/). JSON struck the best balance between
human-editable and machine-editable, allowing both hand-made templates as well
as machine generated templates to easily be made.

We'll start by creating the entire template, then we'll go over each section
briefly. Create a file `example.json` and fill it with the following contents:

``` {.javascript}
{
  "variables": {
    "aws_access_key": "",
    "aws_secret_key": ""
  },
  "builders": [{
    "type": "amazon-ebs",
    "access_key": "{{user `aws_access_key`}}",
    "secret_key": "{{user `aws_secret_key`}}",
    "region": "us-east-1",
    "source_ami": "ami-fce3c696",
    "instance_type": "t2.micro",
    "ssh_username": "ubuntu",
    "ami_name": "packer-example {{timestamp}}"
  }]
}
```

When building, you'll pass in the `aws_access_key` and `aws_secret_key` as a
[user variable](/docs/templates/user-variables.html), keeping your secret keys
out of the template. You can create security credentials on [this
page](https://console.aws.amazon.com/iam/home?#security_credential). An example
IAM policy document can be found in the [Amazon EC2 builder
docs](/docs/builders/amazon.html).

This is a basic template that is ready-to-go. It should be immediately
recognizable as a normal, basic JSON object. Within the object, the `builders`
section contains an array of JSON objects configuring a specific *builder*. A
builder is a component of Packer that is responsible for creating a machine and
turning that machine into an image.

In this case, we're only configuring a single builder of type `amazon-ebs`. This
is the Amazon EC2 AMI builder that ships with Packer. This builder builds an
EBS-backed AMI by launching a source AMI, provisioning on top of that, and
re-packaging it into a new AMI.

The additional keys within the object are configuration for this builder,
specifying things such as access keys, the source AMI to build from, and more.
The exact set of configuration variables available for a builder are specific to
each builder and can be found within the [documentation](/docs).

Before we take this template and build an image from it, let's validate the
template by running `packer validate example.json`. This command checks the
syntax as well as the configuration values to verify they look valid. The output
should look similar to below, because the template should be valid. If there are
any errors, this command will tell you.

``` {.text}
$ packer validate example.json
Template validated successfully.
```

Next, let's build the image from this template.

An astute reader may notice that we said earlier we'd be building an image with
Redis pre-installed, and yet the template we made doesn't reference Redis
anywhere. In fact, this part of the documentation will only cover making a first
basic, non-provisioned image. The next section on provisioning will cover
installing Redis.

## Your First Image

With a properly validated template. It is time to build your first image. This
is done by calling `packer build` with the template file. The output should look
similar to below. Note that this process typically takes a few minutes.

``` {.text}
$ packer build \
    -var 'aws_access_key=YOUR ACCESS KEY' \
    -var 'aws_secret_key=YOUR SECRET KEY' \
    example.json
==> amazon-ebs: amazon-ebs output will be in this color.

==> amazon-ebs: Creating temporary keypair for this instance...
==> amazon-ebs: Creating temporary security group for this instance...
==> amazon-ebs: Authorizing SSH access on the temporary security group...
==> amazon-ebs: Launching a source AWS instance...
==> amazon-ebs: Waiting for instance to become ready...
==> amazon-ebs: Connecting to the instance via SSH...
==> amazon-ebs: Stopping the source instance...
==> amazon-ebs: Waiting for the instance to stop...
==> amazon-ebs: Creating the AMI: packer-example 1371856345
==> amazon-ebs: AMI: ami-19601070
==> amazon-ebs: Waiting for AMI to become ready...
==> amazon-ebs: Terminating the source AWS instance...
==> amazon-ebs: Deleting temporary security group...
==> amazon-ebs: Deleting temporary keypair...
==> amazon-ebs: Build finished.

==> Builds finished. The artifacts of successful builds are:
--> amazon-ebs: AMIs were created:

us-east-1: ami-19601070
```

At the end of running `packer build`, Packer outputs the *artifacts* that were
created as part of the build. Artifacts are the results of a build, and
typically represent an ID (such as in the case of an AMI) or a set of files
(such as for a VMware virtual machine). In this example, we only have a single
artifact: the AMI in us-east-1 that was created.

This AMI is ready to use. If you wanted you can go and launch this AMI right now
and it would work great.

-&gt; **Note:** Your AMI ID will surely be different than the one above. If you
try to launch the one in the example output above, you will get an error. If you
want to try to launch your AMI, get the ID from the Packer output.

## Managing the Image

Packer only builds images. It does not attempt to manage them in any way. After
they're built, it is up to you to launch or destroy them as you see fit. If you
want to store and namespace images for easy reference, you can use [Atlas by
HashiCorp](https://atlas.hashicorp.com). We'll cover remotely building and
storing images at the end of this getting started guide.

After running the above example, your AWS account now has an AMI associated with
it. AMIs are stored in S3 by Amazon, so unless you want to be charged about
$0.01 per month, you'll probably want to remove it. Remove the AMI by first
deregistering it on the [AWS AMI management
page](https://console.aws.amazon.com/ec2/home?region=us-east-1#s=Images). Next,
delete the associated snapshot on the [AWS snapshot management
page](https://console.aws.amazon.com/ec2/home?region=us-east-1#s=Snapshots).

Congratulations! You've just built your first image with Packer. Although the
image was pretty useless in this case (nothing was changed about it), this page
should've given you a general idea of how Packer works, what templates are, and
how to validate and build templates into machine images.
