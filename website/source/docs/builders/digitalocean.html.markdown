---
description: |
    The `digitalocean` Packer builder is able to create new images for use with
    DigitalOcean. The builder takes a source image, runs any provisioning necessary
    on the image after launching it, then snapshots it into a reusable image. This
    reusable image can then be used as the foundation of new servers that are
    launched within DigitalOcean.
layout: docs
page_title: DigitalOcean Builder
...

# DigitalOcean Builder

Type: `digitalocean`

The `digitalocean` Packer builder is able to create new images for use with
[DigitalOcean](https://www.digitalocean.com). The builder takes a source image,
runs any provisioning necessary on the image after launching it, then snapshots
it into a reusable image. This reusable image can then be used as the foundation
of new servers that are launched within DigitalOcean.

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

-   `api_token` (string) - The client TOKEN to use to access your account. It
    can also be specified via environment variable `DIGITALOCEAN_API_TOKEN`,
    if set.

-   `image` (string) - The name (or slug) of the base image to use. This is the
    image that will be used to launch a new droplet and provision it. See
    [https://developers.digitalocean.com/documentation/v2/\#list-all-images](https://developers.digitalocean.com/documentation/v2/#list-all-images) for
    details on how to get a list of the the accepted image names/slugs.

-   `region` (string) - The name (or slug) of the region to launch the
    droplet in. Consequently, this is the region where the snapshot will
    be available. See
    [https://developers.digitalocean.com/documentation/v2/\#list-all-regions](https://developers.digitalocean.com/documentation/v2/#list-all-regions) for
    the accepted region names/slugs.

-   `size` (string) - The name (or slug) of the droplet size to use. See
    [https://developers.digitalocean.com/documentation/v2/\#list-all-sizes](https://developers.digitalocean.com/documentation/v2/#list-all-sizes) for
    the accepted size names/slugs.

### Optional:

-   `api_url` (string) - Non standard api endpoint URL. Set this if you are
    using a DigitalOcean API compatible service. It can also be specified via
    environment variable `DIGITALOCEAN_API_URL`.

-   `droplet_name` (string) - The name assigned to the droplet. DigitalOcean
    sets the hostname of the machine to this value.

-   `private_networking` (boolean) - Set to `true` to enable private networking
    for the droplet being created. This defaults to `false`, or not enabled.

-   `snapshot_name` (string) - The name of the resulting snapshot that will
    appear in your account. This must be unique. To help make this unique, use a
    function like `timestamp` (see [configuration
    templates](/docs/templates/configuration-templates.html) for more info)

-   `state_timeout` (string) - The time to wait, as a duration string, for a
    droplet to enter a desired state (such as "active") before timing out. The
    default state timeout is "6m".

-   `user_data` (string) - User data to launch with the Droplet.

## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
access tokens:

``` {.javascript}
{
  "type": "digitalocean",
  "api_token": "YOUR API KEY",
  "image": "ubuntu-12-04-x64",
  "region": "nyc2",
  "size": "512mb"
}
```
