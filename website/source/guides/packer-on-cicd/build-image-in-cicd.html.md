---
layout: guides
sidebar_current: guides-packer-on-cicd-build-image
page_title: Build Images in CI/CD
---

# Build Images in CI/CD

The following guides from our partners show how to use their services to build
images with Packer.

- [How to Build Immutable Infrastructure with Packer and CircleCI Workflows](https://circleci.com/blog/how-to-build-immutable-infrastructure-with-packer-and-circleci-workflows/)
- [Using Packer and Ansible to Build Immutable Infrastructure in CodeShip](https://blog.codeship.com/packer-ansible/)

The majority of the [Packer Builders](/docs/builders/index.html) can run in
a container or VM, a common model used by most CI/CD services. However, the
[QEMU builder](/docs/builders/qemu.html) for
[KVM](https://www.linux-kvm.org/page/Main_Page) and
[Xen](https://www.xenproject.org/) virtual machine images, [VirtualBox
builder](/docs/builders/virtualbox.html) for OVA or OVF virtual machines and
[VMWare builder](/docs/builders/vmware.html) for use with VMware products
require running on a bare-metal machine.

The [Building a VirtualBox Image with Packer in
TeamCity](/guides/packer-on-cicd/build-virtualbox-image.html) guide shows
how to create a VirtualBox image using TeamCity's support for running scripts
on bare-metal machines.
