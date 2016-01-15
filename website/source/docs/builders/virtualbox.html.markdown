---
description: |
    The VirtualBox Packer builder is able to create VirtualBox virtual machines and
    export them in the OVA or OVF format.
layout: docs
page_title: VirtualBox Builder
...

# VirtualBox Builder

The VirtualBox Packer builder is able to create
[VirtualBox](https://www.virtualbox.org) virtual machines and export them in the
OVA or OVF format.

Packer actually comes with multiple builders able to create VirtualBox machines,
depending on the strategy you want to use to build the image. Packer supports
the following VirtualBox builders:

-   [virtualbox-iso](/docs/builders/virtualbox-iso.html) - Starts from an ISO
    file, creates a brand new VirtualBox VM, installs an OS, provisions software
    within the OS, then exports that machine to create an image. This is best
    for people who want to start from scratch.

-   [virtualbox-ovf](/docs/builders/virtualbox-ovf.html) - This builder imports
    an existing OVF/OVA file, runs provisioners on top of that VM, and exports
    that machine to create an image. This is best if you have an existing
    VirtualBox VM export you want to use as the source. As an additional
    benefit, you can feed the artifact of this builder back into itself to
    iterate on a machine.
