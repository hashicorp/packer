---
description: |
    HyperOne Packer builder creates new images on the HyperOne platform.
    The builder takes a source image, runs any provisioning necessary on
    the image after launching it, then creates a reusable image.
layout: docs
page_title: 'HyperOne - Builders'
sidebar_current: 'docs-builders-hyperone'
---

# HyperOne Builder

Type: `hyperone`

The `hyperone` Packer builder is able to create new images on the [HyperOne
platform](http://www.hyperone.com/). The builder takes a source image, runs
any provisioning necessary on the image after launching it, then creates a
reusable image.

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

-   `disk_size` (float) - Size of the created disk, in GiB.

-   `project` (string) - The id or name of the project. This field is required
    only if using session tokens. It should be skipped when using service
    account authentication.

-   `source_image` (string) - ID or name of the image to launch server from.

-   `token` (string) - The authentication token used to access your account.
    This can be either a session token or a service account token.
    If not defined, the builder will attempt to find it in the following order:

    - In `HYPERONE_TOKEN` environment variable.
    - In `~/.h1-cli/conf.json` config file used by [h1-cli](https://github.com/hyperonecom/h1-cli).
    - By using SSH authentication if `token_login` variable has been set.

-   `vm_flavour` (string) - ID or name of the type this server should be created with.

### Optional:

-   `disk_name` (string) - The name of the created disk.

-   `disk_type` (string) - The type of the created disk.

-   `image_description` (string) - The description of the resulting image.

-   `image_name` (string) - The name of the resulting image. Defaults to
    "packer-{{timestamp}}"
    (see [configuration templates](/docs/templates/engine.html) for more info).

-   `image_service` (string) - The service of the resulting image.

-   `image_tags` (map of key/value strings) - Key/value pair tags to
    add to the created image.

-   `network` (string) - The ID of the network to attach to the created server.

-   `private_ip` (string) - The ID of the private IP within chosen `network`
    that should be assigned to the created server.

-   `public_ip` (string) - The ID of the public IP that should be assigned to
    the created server. If `network` is chosen, the public IP will be associated
    with server's private IP.

-   `ssh_keys` (array of strings) - List of SSH keys by name or id to be added
    to the server on launch.

-   `state_timeout` (string) - Timeout for waiting on the API to complete
    a request. Defaults to 5m.

-   `token_login` (string) - Login (an e-mail) on HyperOne platform. Set this
    if you want to fetch the token by SSH authentication.

-   `user_data` (string) - User data to launch with the server. Packer will not
    automatically wait for a user script to finish before shutting down the
    instance, this must be handled in a provisioner.

-   `vm_name` (string) - The name of the created server.

-   `vm_tags` (map of key/value strings) - Key/value pair tags to
    add to the created server.

## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
token.

``` json
{
  "type": "hyperone",
  "token": "YOUR_AUTH_TOKEN",
  "source_image": "ubuntu-18.04",
  "vm_flavour": "a1.nano",
  "disk_size": 10
}
```
