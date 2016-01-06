---
description: |
    The VMware Packer builder is able to create VMware virtual machines for use with
    any VMware product.
layout: docs
page_title: VMware Builder
...

# VMware Builder

The VMware Packer builder is able to create VMware virtual machines for use with
any VMware product.

Packer actually comes with multiple builders able to create VMware machines,
depending on the strategy you want to use to build the image. Packer supports
the following VMware builders:

-   [vmware-iso](/docs/builders/vmware-iso.html) - Starts from an ISO file,
    creates a brand new VMware VM, installs an OS, provisions software within
    the OS, then exports that machine to create an image. This is best for
    people who want to start from scratch.

-   [vmware-vmx](/docs/builders/vmware-vmx.html) - This builder imports an
    existing VMware machine (from a VMX file), runs provisioners on top of that
    VM, and exports that machine to create an image. This is best if you have an
    existing VMware VM you want to use as the source. As an additional benefit,
    you can feed the artifact of this builder back into Packer to iterate on
    a machine.
