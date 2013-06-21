---
layout: "docs"
---

# Packer Documentation

Welcome to the Packer documentation! This documentation will guide you from
complete beginner to being a Packer expert. It introduces all the concepts
of Packer as well as contains references material for every configuration
parameter and command-line flags available to control Packer.

## What is Packer?

Packer is a tool for creating identical machine images for multiple platforms
from a single source configuration. Packer is lightweight, runs on every major
operating system, and is highly performant, creating machine images for
multiple platforms in parallel. Packer does not replace configuration management
like Chef or Puppet. In fact, when building images, Packer is able to use tools
like Chef or Puppet to install software onto the image.

A _machine image_ is a single static unit that contains a pre-configured operating
system and installed software which is used to quickly create new running machines.
Machine image formats change for each platform. Some examples include
[AMIs](http://en.wikipedia.org/wiki/Amazon_Machine_Image) for EC2,
VMDK/VMX files for VMware, OVF exports for VirtualBox, etc.

## Why Use Packer?

Historically, creating
these images has been a predominantly manual process. Any existing automated tools were able to
create only one type of image. Packer, on the other hand, is able to automatically
create any type of image, all from a single source configuration. This unlocks
untapped potential in developing, testing, and deploying applications.

Pre-baked machine images have a lot of advantages, but we've been unable to
benefit from them because they have been too tedious to create and manage.
Packer tears down this barrier, allowing the benefits of pre-baked machine
images to become available to everyone. Some benefits include:

* ***Super fast infrastructure deployment***. Packer images allow you to launch
completely provisioned and configured machines in seconds, rather than
several minutes or hours. This benefits not only production, but development as well,
since development virtual machines can also be launched in seconds, without waiting
for a typically much longer provisioning time.

* ***Multi-provider portability***. Because Packer creates identical images for
multiple platforms, you can run production in AWS, staging/QA in a private
cloud like OpenStack, and development in desktop virtualization solutions
such as VMware or VirtualBox. Each environment is running an identical
machine image, giving ultimate portability.

* ***Improved stability***. Packer installs and configures all the software for
a machine at the time the image is built. If there are bugs in these scripts,
they'll be caught early, rather than several minutes after a machine is launched.

* ***Greater testability***. After a machine image is built, that machine image
can be quickly launched and smoke tested to verify that things appear to be
working. If they are, you can be confident that any other machines launched
from that image will function properly.

Packer makes it extremely easy to take advantage of all these benefits.

What are you waiting for? Let's get started!<D-j>
