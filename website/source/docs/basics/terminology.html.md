---
description: |
    There are a handful of terms used throughout the Packer documentation where the
    meaning may not be immediately obvious if you haven't used Packer before.
    Luckily, there are relatively few. This page documents all the terminology
    required to understand and use Packer. The terminology is in alphabetical order
    for easy referencing.
layout: docs
page_title: Packer Terminology
...

# Packer Terminology

There are a handful of terms used throughout the Packer documentation where the
meaning may not be immediately obvious if you haven't used Packer before.
Luckily, there are relatively few. This page documents all the terminology
required to understand and use Packer. The terminology is in alphabetical order
for easy referencing.

-   `Artifacts` are the results of a single build, and are usually a set of IDs
    or files to represent a machine image. Every builder produces a
    single artifact. As an example, in the case of the Amazon EC2 builder, the
    artifact is a set of AMI IDs (one per region). For the VMware builder, the
    artifact is a directory of files comprising the created virtual machine.

-   `Builds` are a single task that eventually produces an image for a
    single platform. Multiple builds run in parallel. Example usage in a
    sentence: "The Packer build produced an AMI to run our web application." Or:
    "Packer is running the builds now for VMware, AWS, and VirtualBox."

-   `Builders` are components of Packer that are able to create a machine image
    for a single platform. Builders read in some configuration and use that to
    run and generate a machine image. A builder is invoked as part of a build in
    order to create the actual resulting images. Example builders include
    VirtualBox, VMware, and Amazon EC2. Builders can be created and added to
    Packer in the form of plugins.

-   `Commands` are sub-commands for the `packer` program that perform some job.
    An example command is "build", which is invoked as `packer build`. Packer
    ships with a set of commands out of the box in order to define its
    command-line interface. Commands can also be created and added to Packer in
    the form of plugins.

-   `Post-processors` are components of Packer that take the result of a builder
    or another post-processor and process that to create a new artifact.
    Examples of post-processors are compress to compress artifacts, upload to
    upload artifacts, etc.

-   `Provisioners` are components of Packer that install and configure software
    within a running machine prior to that machine being turned into a
    static image. They perform the major work of making the image contain
    useful software. Example provisioners include shell scripts, Chef,
    Puppet, etc.

-   `Templates` are JSON files which define one or more builds by configuring
    the various components of Packer. Packer is able to read a template and use
    that information to create multiple machine images in parallel.
