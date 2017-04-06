---
layout: docs
sidebar_current: docs-builders-scaleway
page_title: Scaleway - Builders
description: |-
  The Scaleway Packer builder is able to create new snapshots for use with
  Scaleway BareMetal and Virtual cloud server. The builder takes a source image, runs any provisioning
  necessary on the image after launching it, then snapshots it into a reusable
  image. This reusable image can then be used as the foundation of new servers
  that are launched within Scaleway.
---


# Scaleway Builder

Type: `scaleway`

The `scaleway` Packer builder is able to create new snapshots for use with
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

- `api_organization` (string) - The organization ID to use to access your account.
    It can also be specified via
    environment variable `SCALEWAY_API_ORGANIZATION`.

- `api_token` (string) - The organization TOKEN to use to access your account.
    It can also be specified via
    environment variable `SCALEWAY_API_TOKEN`.

- `image` (string) - The UUID of the base image to use. This is the
    image that will be used to launch a new server and provision it. See
    [https://api-marketplace.scaleway.com/images](https://api-marketplace.scaleway.com/images) to
    get the complete list of the accepted image UUID.

- `region` (string) - The name of the region to launch the
    server in (`par1` or `ams1`). Consequently, this is the region where the snapshot will
    be available.

- `commercial_type` (string) - The name of the server commercial type: `C1`, `C2S`, `C2M`,
    `C2L`, `X64-2GB`, `X64-4GB`, `X64-8GB`, `X64-15GB`, `X64-30GB`, `X64-60GB`, `X64-120GB`

### Optional:

- `server_name` (string) - The name assigned to the server.

- `snapshot_name` (string) - The name of the resulting snapshot that will
    appear in your account.

## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
access tokens:

```json
{
  "type": "scaleway",
  "api_organization": "YOUR ORGANIZATION KEY",
  "api_token": "YOUR TOKEN",
  "image": "f01f8a48-c026-48ac-9771-a70eaac0890e",
  "region": "par1",
  "commercial_type": "X64-2GB",
  "ssh_username": "root"
}
```
