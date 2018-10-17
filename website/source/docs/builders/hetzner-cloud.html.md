---
description: |
    The Hetzner Cloud Packer builder is able to create new images for use with
    DigitalOcean. The builder takes a source image, runs any provisioning
    necessary on the image after launching it, then snapshots it into a reusable
    image. This reusable image can then be used as the foundation of new servers
    that are launched within the Hetzner Cloud.
layout: docs
page_title: 'Hetzner Cloud - Builders'
sidebar_current: 'docs-builders-hetzner-cloud'
---

# Hetzner Cloud Builder

Type: `hcloud`

The `hcloud` Packer builder is able to create new images for use with
[Hetzner Cloud](https://www.hetzner.cloud). The builder takes a source image,
runs any provisioning necessary on the image after launching it, then snapshots
it into a reusable image. This reusable image can then be used as the foundation
of new servers that are launched within the Hetzner Cloud.

The builder does *not* manage images. Once it creates an image, it is up to you
to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `token` (string) - The client TOKEN to use to access your account. It
    can also be specified via environment variable `HCLOUD_TOKEN`,
    if set.

-   `image` (string) - 	ID or name of location to create server in. 

-   `location` (string) - The name of the location to launch the
    droplet in.

-   `server_type` (string) - ID or name of the server type this server should be created with.

### Optional:

-   `endpoint` (string) - Non standard api endpoint URL. Set this if you are
    using a DigitalOcean API compatible service. It can also be specified via
    environment variable `HCLOUD_ENDPOINT`.

-   `server_name` (string) - The name assigned to the droplet. DigitalOcean
    sets the hostname of the machine to this value.

-   `snapshot_name` (string) - The name of the resulting snapshot that will
    appear in your account. Defaults to "packer-{{timestamp}}" (see
    [configuration templates](/docs/templates/engine.html) for more info).

-   `poll_interval` (string) - Configures the interval in which actions are polled by the client. Default `500ms`. Increase this interval if you run into rate limiting errors.

-   `user_data` (string) - User data to launch with the Droplet.

-   `user_data_file` (string) - Path to a file that will be used for the user
    data when launching the Droplet.

-   `tags` (list) - Tags to apply to the droplet when it is created

## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
access tokens:

``` json
{
  "type": "hcloud",
  "token": "YOUR API KEY",
  "image": "ubuntu-18.04",
  "location": "nbg1",
  "server_type": "cx11",
  "ssh_username": "root"
}
```
