---
description: |
  Packer is able to create custom images using Oracle Cloud Infrastructure.
layout: docs
page_title: 'Oracle - Builders'
sidebar_current: 'docs-builders-oracle'
---

# Oracle Builder

Packer is able to create custom images on both Oracle Cloud Infrastructure and
Oracle Cloud Infrastructure Classic Compute. Packer comes with builders
designed to support both platforms. Please choose the one that's right for you:

-   [oracle-classic](/docs/builders/oracle-classic.html) - Create custom images
    in Oracle Cloud Infrastructure Classic Compute by launching a source
    instance and creating an image list from a snapshot of it after
    provisioning.

-   [oracle-oci](/docs/builders/oracle-oci.html) - Create custom images in
    Oracle Cloud Infrastructure (OCI) by launching a base instance and creating
    an image from it after provisioning.
