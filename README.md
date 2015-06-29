# Packer

[![Build Status](https://travis-ci.org/mitchellh/packer.svg?branch=master)](https://travis-ci.org/mitchellh/packer)
[![Windows Build Status](https://ci.appveyor.com/api/projects/status/github/mitchellh/packer?branch=master&svg=true)](https://ci.appveyor.com/project/hashicorp/packer)

* Website: http://www.packer.io
* IRC: `#packer-tool` on Freenode
* Mailing list: [Google Groups](http://groups.google.com/group/packer-tool)

Packer is a tool for building identical machine images for multiple platforms
from a single source configuration.

Packer is lightweight, runs on every major operating system, and is highly
performant, creating machine images for multiple platforms in parallel.
Packer comes out of the box with support for the following platforms:
* Amazon EC2 (AMI). Both EBS-backed and instance-store AMIs
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
file as `quick-start.json`. Be sure to replace any credentials with your
own.

```json
{
  "builders": [{
    "type": "amazon-ebs",
    "access_key": "YOUR KEY HERE",
    "secret_key": "YOUR SECRET KEY HERE",
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

Full, comprehensive documentation is viewable on the Packer website:

http://www.packer.io/docs

## Developing Packer

If you wish to work on Packer itself or any of its built-in providers,
you'll first need [Go](http://www.golang.org) installed (version 1.4+ is
_required_). Make sure Go is properly installed, including setting up
a [GOPATH](http://golang.org/doc/code.html#GOPATH).

Next, install the following software packages, which are needed for some dependencies:

- [Bazaar](http://bazaar.canonical.com/en/)
- [Git](http://git-scm.com/)
- [Mercurial](http://mercurial.selenic.com/)

Then, install [Gox](https://github.com/mitchellh/gox), which is used
as a compilation tool on top of Go:

    $ go get -u github.com/mitchellh/gox

Next, clone this repository into `$GOPATH/src/github.com/mitchellh/packer`.
Install the necessary dependencies by running `make updatedeps` and then just
type `make`. This will compile some more dependencies and then run the tests. If
this exits with exit status 0, then everything is working!

    $ make updatedeps
    ...
    $ make
    ...

To compile a development version of Packer and the built-in plugins,
run `make dev`. This will put Packer binaries in the `bin` folder:

    $ make dev
    ...
    $ bin/packer
    ...


If you're developing a specific package, you can run tests for just that
package by specifying the `TEST` variable. For example below, only
`packer` package tests will be run.

    $ make test TEST=./packer
    ...

### Acceptance Tests

Packer has comprehensive [acceptance tests](https://en.wikipedia.org/wiki/Acceptance_testing)
covering the builders of Packer.

If you're working on a feature of a builder or a new builder and want
verify it is functioning (and also hasn't broken anything else), we recommend
running the acceptance tests.

**Warning:** The acceptance tests create/destroy/modify *real resources*, which
may incur real costs in some cases. In the presence of a bug, it is technically
possible that broken backends could leave dangling data behind. Therefore,
please run the acceptance tests at your own risk. At the very least,
we recommend running them in their own private account for whatever builder
you're testing.

To run the acceptance tests, invoke `make testacc`:

```sh
$ make testacc TEST=./builder/amazon/ebs
...
```

The `TEST` variable is required, and you should specify the folder where the
backend is. The `TESTARGS` variable is recommended to filter down to a specific
resource to test, since testing all of them at once can sometimes take a very
long time.

Acceptance tests typically require other environment variables to be set for
things such as access keys. The test itself should error early and tell
you what to set, so it is not documented here.
