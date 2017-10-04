---
description: |
    The Hyper-V Packer builder is able to create Hyper-V virtual machines and
    export them.
layout: docs
page_title: 'Hyper-V - Builders'
sidebar_current: 'docs-builders-hyperv'
---

# HyperV Builder

The HyperV Packer builder is able to create [Hyper-V](https://www.microsoft.com/en-us/server-cloud/solutions/virtualization.aspx)
virtual machines and export them.

-   [hyperv-iso](/docs/builders/hyperv-iso.html) - Starts from
    an ISO file, creates a brand new Hyper-V VM, installs an OS,
    provisions software within the OS, then exports that machine to create
    an image. This is best for people who want to start from scratch.

-   [hyperv-vmcx](/docs/builders/hyperv-vmcx.html) - Clones an
    an existing virtual machine, provisions software within the OS, 
    then exports that machine to create an image. This is best for 
    people who have existing base images and want to customize them.