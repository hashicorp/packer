---
description: |
    The `brkt` Packer builder is able to create Bracket Image Definitions
    for use with the Bracket Computing Cell. The builder takes a source image,
    runs any provisioning necessary on the image after launching it, then
    snapshots it into a reusable image. This reusable image can then be used as
    the foundation of new servers that are launched within Bracket.
layout: docs
page_title: 'Bracket Image Definition Builder'
...

# Bracket Image Definition Builder

Type `brkt`

The `brkt` Packer builder is able to create new images for use with
[DigitalOcean](http://www.digitalocean.com). The builder takes a source image,
runs any provisioning necessary on the image after launching it, then
snapshots it into a reusable image. This reusable image can then be used as
the foundation of new servers that are launched within Bracket.

The builder does *not* manage images. Once it creates an image, it is up to you
to use it or delete it.


## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) should be configured for this
builder. The minimal configuration needs `communicator` and `*_username` to be
set.

### Required:

-   `access_token` (string) - The access token used to communicate with Bracket.

-   `billing_group_uuid` (string) - The Billing Group that will be charged for
    launching the provisioning instance.

-   `computing_cell` (string) - The Computing Cell that the provisioning
    instance will be launched into.

-   `image_definition_uuid` (string) - The initial Image Definition used as a
    base for the newly created machine.

-   `image_name` (string) - The name of the resulting Image Definition that
    will appear when managing Image Definitions in the Bracket console or via
    APIs. This must be unique. To help make this unique, use a function like
    `timestamp` (see [configuration templates](/docs/templates/configuration-templates.html) for more info)

-   `mac_key` (string) - The MAC key used to communicate with Bracket.

-   `min_cpu_cores` (integer) - Specify the minimum amount of CPU cores needed
    for the instance that will be provisioning the image. This will be combined
    with the `min_ram_in_gb` to find a matching `machine_type`.

-   `min_ram_in_gb` (integer) - Specify the minimum amount of RAM needed for
    the instance that will be provisioning the image. This will be combined
    with the `min_cpu_cores` to find a matching `machine_type`.


### Optional:

-   `cloud_config` (object) - Cloud-Config settings.
    [Learn more about Cloud-Config.](https://coreos.com/os/docs/latest/cloud-config.html)

-   `machine_type_uuid` (string) - Machine Type to use for the instance, this
    is an alternative to using `min_cpu_cores` and `min_ram_in_gb`.

-   `metavisor_enabled` (boolean) - Enable Metavisor on the deployed instance.
    This is required if you are deploying an encrypted Image Definition.

-   `security_group_uuid` (string) - The Security Group to launch this instance
    in.

-   `zone_uuid` (string) - The UUID of the Zone to launch the provisioning
    instance in.


## Basic Example

Here is a basic example. It is completely valid except for the access keys and
the UUIDs which are unique for your account:

``` {.javascript}
{
  "type": "brkt",

  "access_token": "YOUR KEY HERE",
  "mac_key": "YOUR MAC KEY HERE",

  "image_definition_uuid": "YOUR IMAGE DEFINITION UUID HERE",
  "billing_group_uuid": "YOUR BILLING GROUP UUID HERE",
  "computing_cell_uuid": "YOUR COMPUTING CELL UUID HERE",
  "security_group_uuid": "YOUR SECURITY GROUP UUID HERE",

  "min_ram_in_gb": 4,
  "min_cpu_cores": 2,

  "image_name": "packer-brkt {{timestamp}}",

  "communicator": "ssh",
  "ssh_username": "ubuntu" // if using ubuntu
}
```