---
layout: guides
sidebar_current: guides-packer-on-cicd-build-image
page_title: Building Images in CI/CD
description: |-
  ...
---

# Building Images in CI/CD

The following guides from our amazing partners show how to use their service to build images with Packer.

- [How to Build Immutable Infrastructure with Packer and CircleCI Workflows](https://docs.google.com/document/d/1hetlS94SpUQ979K-1At9hwn1oDNWHegBqHGNcIFLBoY/edit)
- [Using Packer and Ansible to Build Immutable Infrastructure](https://blog.codeship.com/packer-ansible/)

For the majority of the [Packer Builders](https://www.packer.io/docs/builders/index.html) can run in a container or VM, a common model used by most CI/CD services. However, the [QEMU builder](https://www.packer.io/docs/builders/qemu.html) for [KVM](https://www.linux-kvm.org/page/Main_Page) and [Xen](https://www.xenproject.org/) virtual machine images, [VirtualBox builder](https://www.packer.io/docs/builders/virtualbox.html) for OVA or OVF virtual machines and [VMWare builder](https://www.packer.io/docs/builders/vmware.html) for use with VMware products require running on a bare-metal machine.

[Building a VirtualBox Image with Packer in TeamCity](https://docs.google.com/document/d/1AQjn4PpApnZ6pf097HYZzZa4ZMspRATxo9wNj78hLLc/edit#)
