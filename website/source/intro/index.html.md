---
description: |
    Welcome to the world of Packer! This introduction guide will show you what
    Packer is, explain why it exists, the benefits it has to offer, and how you can
    get started with it. If you're already familiar with Packer, the documentation
    provides more of a reference for all available features.
layout: intro
next_title: 'Why Use Packer?'
next_url: '/intro/why.html'
page_title: Introduction
prev_url: '# '
...

# Introduction to Packer

Welcome to the world of Packer! This introduction guide will show you what
Packer is, explain why it exists, the benefits it has to offer, and how you can
get started with it. If you're already familiar with Packer, the
[documentation](/docs) provides more of a reference for all available features.

## What is Packer?

Packer is an open source tool for creating identical machine images for multiple
platforms from a single source configuration. Packer is lightweight, runs on
every major operating system, and is highly performant, creating machine images
for multiple platforms in parallel. Packer does not replace configuration
management like Chef or Puppet. In fact, when building images, Packer is able to
use tools like Chef or Puppet to install software onto the image.

A *machine image* is a single static unit that contains a pre-configured
operating system and installed software which is used to quickly create new
running machines. Machine image formats change for each platform. Some examples
include [AMIs](https://en.wikipedia.org/wiki/Amazon_Machine_Image) for EC2,
VMDK/VMX files for VMware, OVF exports for VirtualBox, etc.
