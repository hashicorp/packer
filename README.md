# Packer

[![Build Status][circleci-badge]][circleci]
[![Discuss](https://img.shields.io/badge/discuss-packer-3d89ff?style=flat)](https://discuss.hashicorp.com/c/packer)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/hashicorp/packer)](https://pkg.go.dev/github.com/hashicorp/packer)
[![GoReportCard][report-badge]][report]
[![codecov](https://codecov.io/gh/hashicorp/packer/branch/master/graph/badge.svg)](https://codecov.io/gh/hashicorp/packer)

[circleci-badge]: https://circleci.com/gh/hashicorp/packer.svg?style=svg
[circleci]: https://app.circleci.com/pipelines/github/hashicorp/packer
[appveyor-badge]: https://ci.appveyor.com/api/projects/status/miavlgnp989e5obc/branch/master?svg=true
[godoc-badge]: https://godoc.org/github.com/hashicorp/packer?status.svg
[godoc]: https://godoc.org/github.com/hashicorp/packer	
[report-badge]: https://goreportcard.com/badge/github.com/hashicorp/packer	
[report]: https://goreportcard.com/report/github.com/hashicorp/packer

<p align="center" style="text-align:center;">
  <a href="https://www.packer.io">
    <img alt="HashiCorp Packer logo" src="website/public/img/logo-packer-padded.svg" width="500" />
  </a>
</p>

Packer is a tool for building identical machine images for multiple platforms
from a single source configuration.

Packer is lightweight, runs on every major operating system, and is highly
performant, creating machine images for multiple platforms in parallel. Packer
comes out of the box with support for many platforms, the full list of which can
be found at https://www.packer.io/docs/builders.

Support for other platforms can be added via plugins.

The images that Packer creates can easily be turned into
[Vagrant](http://www.vagrantup.com) boxes.

## Quick Start

**Note:** There is a great
[introduction and getting started guide](https://www.packer.io/intro)
for those with a bit more patience. Otherwise, the quick start below
will get you up and running quickly, at the sacrifice of not explaining some
key points.

First, [download a pre-built Packer
binary](https://www.packer.io/downloads.html) for your operating system or
[compile Packer
yourself](https://github.com/hashicorp/packer/blob/master/.github/CONTRIBUTING.md#setting-up-go-to-work-on-packer).

After Packer is installed, create your first template, which tells Packer
what platforms to build images for and how you want to build them. In our
case, we'll create a simple AMI that has Redis pre-installed. 

Save this file as `quick-start.pkr.hcl`. Export your AWS credentials as the
`AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` environment variables.

```hcl
variable "access_key" {
  type    = string
  default = "${env("AWS_ACCESS_KEY_ID")}"
}

variable "secret_key" {
  type    = string
  default = "${env("AWS_SECRET_ACCESS_KEY")}"
}

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "amazon-ebs" "quick-start" {
  access_key    = "${var.access_key}"
  ami_name      = "packer-example ${local.timestamp}"
  instance_type = "t2.micro"
  region        = "us-east-1"
  secret_key    = "${var.secret_key}"
  source_ami    = "ami-af22d9b9"
  ssh_username  = "ubuntu"
}

build {
  sources = ["source.amazon-ebs.quick-start"]
}
```

Next, tell Packer to build the image:

```
$ packer build quick-start.pkr.hcl
...
```

Packer will build an AMI according to the "quick-start" template. The AMI
will be available in your AWS account. To delete the AMI, you must manually
delete it using the [AWS console](https://console.aws.amazon.com/). Packer
builds your images, it does not manage their lifecycle. Where they go, how
they're run, etc., is up to you.

## Documentation

Comprehensive documentation is viewable on the Packer website at https://www.packer.io/docs.

## Contributing to Packer

See
[CONTRIBUTING.md](https://github.com/hashicorp/packer/blob/master/.github/CONTRIBUTING.md)
for best practices and instructions on setting up your development environment
to work on Packer.

## Unmaintained Plugins
As contributors' circumstances change, development on a community maintained
plugin can slow. When this happens, the Packer team may mark a plugin as
unmaintained, to clearly signal the plugin's status to users.

What does **unmaintained** mean?

1. The code repository and all commit history will still be available.
1. Documentation will remain on the Packer website.
1. Issues and pull requests are monitored as a best effort.
1. No active development will be performed by the Packer team.

If anyone form them community is interested in maintaining a community
supported plugin, please feel free to submit contributions via a pull-
request for review; reviews are generally prioritized over feature work
when possible. For a list of open plugin issues and pending feature requests see the [Packer Issue Tracker](https://github.com/hashicorp/packer/issues/).
