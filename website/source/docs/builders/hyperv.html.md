---
layout: "docs"
page_title: "Hyper-V Builder"
description: |-
  The Hyper-V Packer builder is able to create Hyper-V virtual machines and export them.
---

# HyperV Builder

The HyperV Packer builder is able to create [Hyper-V](https://www.microsoft.com/en-us/server-cloud/solutions/virtualization.aspx)
virtual machines and export them.

Packer currently only support building HyperV machines with an iso:

* [hyperv-iso](/docs/builders/hyperv-iso.html) - Starts from
  an ISO file, creates a brand new Hyper-V VM, installs an OS,
  provisions software within the OS, then exports that machine to create
  an image. This is best for people who want to start from scratch.
