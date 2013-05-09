# Packer

Packer is a tool for building identical machine images across multiple clouds.

Packer provides a framework and configuration format for creating identical
machine images to launch into any environment, such as VirtualBox, VMware,
Amazon EC2, etc. Because this build process is automated, you can develop in
VirtualBox, then deploy to EC2 with an identical image.

## Quick Start

First, get Packer by either downloading a pre-built Packer binary for
your operating system or [downloading and compiling Packer yourself](#developing-packer).

After Packer is installed, build your first machine image.

```
$ packer build quick-start.json
...
```

Packer will build an AMI according to the "quick-start" template. The AMI
will be available in your AWS account. To delete the AMI, you must manually
delete it using the [AWS console](https://console.aws.amazon.com/). Packer
builds your images, it does not manage their lifecycle. Where they go, how
they're run, etc. is up to you.

## Templates

Templates are static configurations that describe what machine images
you want to create, how to create them, and what format you finally want
them to be in.

Packer reads a template and builds all the requested machine images
in parallel.

Templates are written in [TOML](https://github.com/mojombo/toml). TOML is
a fantastic configuration language that you can learn in minutes, and is
very human-readable as well.

First, a complete template is shown below. Then, the details and
structure of a template are discussed:

```toml
name = "my-custom-image"

[builder.amazon-ebs]
source = "ami-de0d9eb7"

[provision]

  [provison.shell]
  type = "shell"
  path = "script.sh"

[output]

  [output.vagrant]
```

Templates are comprised of three parts:

* **builders** (1 or more) specify how the initial running system is
  built.

* **provisioners** (0 or more) specify how to install and configure
  software from within the base running system.

* **outputs** (0 or more) specify what to do with the completed system.
  For example, these can output [Vagrant](http://www.vagrantup.com)-compatible
  boxes, gzipped files, etc.

## Developing Packer

If you wish to work on Packer itself, you'll first need [Go](http://golang.org)
installed (version 1.1+ is _required_). Next, clone this repository then just type `make`.
In a few moments, you'll have a working `packer` executable:

```
$ make
...
$ bin/packer
...
```

You can run tests by typing `make test`. This will run tests for Packer core
along with all the core builders and commands and such that come with Packer.
