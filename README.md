# Packer

[![Build Status](https://travis-ci.org/mitchellh/packer.svg?branch=master)](https://travis-ci.org/mitchellh/packer)
[![Windows Build Status](https://ci.appveyor.com/api/projects/status/github/mitchellh/packer?branch=master&svg=true)](https://ci.appveyor.com/project/hashicorp/packer)

* Website: http://www.packer.io
* IRC: `#packer-tool` on Freenode
* Mailing list: [Google Groups](http://groups.google.com/group/packer-tool)

Packer is a tool for building identical machine images for multiple platforms
from a single source configuration.

Packer is lightweight, runs on every major operating system, and is highly
performant, creating machine images for multiple platforms in parallel. Packer
comes out of the box with support for the following platforms:

* Amazon EC2 (AMI). Both EBS-backed and instance-store AMIs
* Azure
* DigitalOcean
* Docker
* Google Compute Engine
* OpenStack
* Parallels
* QEMU. Both KVM and Xen images.
* VirtualBox
* VMware

Support for other platforms can be added via plugins.

The images that Packer creates can easily be turned into
[Vagrant](http://www.vagrantup.com) boxes.

## Quick Start

**Note:** There is a great
[introduction and getting started guide](http://www.packer.io/intro)
for those with a bit more patience. Otherwise, the quick start below
will get you up and running quickly, at the sacrifice of not explaining some
key points.

First, [download a pre-built Packer binary](http://www.packer.io/downloads.html)
for your operating system or [compile Packer yourself](#developing-packer).

After Packer is installed, create your first template, which tells Packer
what platforms to build images for and how you want to build them. In our
case, we'll create a simple AMI that has Redis pre-installed. Save this
file as `quick-start.json`. Export your AWS credentials as the
`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.

```json
{
  "variables": {
    "access_key": "{{env `AWS_ACCESS_KEY_ID`}}",
    "secret_key": "{{env `AWS_SECRET_ACCESS_KEY`}}"
  },
  "builders": [{
    "type": "amazon-ebs",
    "access_key": "{{user `access_key`}}",
    "secret_key": "{{user `secret_key`}}",
    "region": "us-east-1",
    "source_ami": "ami-de0d9eb7",
    "instance_type": "t1.micro",
    "ssh_username": "ubuntu",
    "ami_name": "packer-example {{timestamp}}"
  }]
}
```

Next, tell Packer to build the image:

```
$ packer build quick-start.json
...
```

Packer will build an AMI according to the "quick-start" template. The AMI
will be available in your AWS account. To delete the AMI, you must manually
delete it using the [AWS console](https://console.aws.amazon.com/). Packer
builds your images, it does not manage their lifecycle. Where they go, how
they're run, etc. is up to you.

## Documentation

Comprehensive documentation is viewable on the Packer website:

http://www.packer.io/docs

## Developing Packer

See [CONTRIBUTING.md](https://github.com/mitchellh/packer/blob/master/CONTRIBUTING.md) for best practices and instructions on setting up your development environment to work on Packer.
