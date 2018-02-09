---
layout: docs
sidebar_current: docs-builders-scaleway
page_title: Scaleway - Builders
description: |-
  The Scaleway Packer builder is able to create new images for use with
  Scaleway BareMetal and Virtual cloud server. The builder takes a source
  image, runs any provisioning necessary on the image after launching it, then
  snapshots it into a reusable image. This reusable image can then be used as
  the foundation of new servers that are launched within Scaleway.
---


# Scaleway Builder

Type: `scaleway`

The `scaleway` Packer builder is able to create new images for use with
[Scaleway](https://www.scaleway.com). The builder takes a source image,
runs any provisioning necessary on the image after launching it, then snapshots
it into a reusable image. This reusable image can then be used as the foundation
of new servers that are launched within Scaleway.

The builder does *not* manage snapshots. Once it creates an image, it is up to you
to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `api_access_key` (string) - The api\_access\_key to use to access your
    account. It can also be specified via environment variable
    `SCALEWAY_API_ACCESS_KEY`. Your access key is available in the
    ["Credentials" section](https://cloud.scaleway.com/#/credentials) of the
    control panel.

-   `api_token` (string) - The organization TOKEN to use to access your
    account. It can also be specified via environment variable
    `SCALEWAY_API_TOKEN`. Your tokens are available in the ["Credentials"
    section](https://cloud.scaleway.com/#/credentials) of the control panel.

-   `image` (string) - The UUID of the base image to use. This is the image
    that will be used to launch a new server and provision it. See
    <https://api-marketplace.scaleway.com/images> get the complete list of the
    accepted image UUID.

-   `region` (string) - The name of the region to launch the server in (`par1`
    or `ams1`). Consequently, this is the region where the snapshot will be
    available.

-   `commercial_type` (string) - The name of the server commercial type: `C1`,
    `C2S`, `C2M`, `C2L`, `X64-2GB`, `X64-4GB`, `X64-8GB`, `X64-15GB`,
    `X64-30GB`, `X64-60GB`, `X64-120GB`, `ARM64-2GB`, `ARM64-4GB`, `ARM64-8GB`,
    `ARM64-16GB`, `ARM64-32GB`, `ARM64-64GB`, `ARM64-128GB`

### Optional:

-   `server_name` (string) - The name assigned to the server. Default
    `packer-UUID`

-   `image_name` (string) - The name of the resulting image that will appear in
    your account. Default `packer-TIMESTAMP`

-   `snapshot_name` (string) - The name of the resulting snapshot that will
    appear in your account. Default `packer-TIMESTAMP`

## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
access tokens:

```json
{
  "type": "scaleway",
  "api_access_key": "YOUR API ACCESS KEY",
  "api_token": "YOUR TOKEN",
  "image": "UUID OF THE BASE IMAGE",
  "region": "par1",
  "commercial_type": "X64-2GB",
  "ssh_username": "root",
  "ssh_private_key_file": "~/.ssh/id_rsa"
}
```

When you do not specified the `ssh_private_key_file`, a temporarily SSH keypair
is generated to connect the server. This key will only allow the `root` user to
connect the server.
