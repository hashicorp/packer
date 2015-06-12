---
layout: "docs"
page_title: "DigitalOcean Builder"
description: |-
  The `digitalocean` Packer builder is able to create new images for use with DigitalOcean. The builder takes a source image, runs any provisioning necessary on the image after launching it, then snapshots it into a reusable image. This reusable image can then be used as the foundation of new servers that are launched within DigitalOcean.
---

# DigitalOcean Builder

Type: `digitalocean`

The `digitalocean` Packer builder is able to create new images for use with
[DigitalOcean](http://www.digitalocean.com). The builder takes a source
image, runs any provisioning necessary on the image after launching it,
then snapshots it into a reusable image. This reusable image can then be
used as the foundation of new servers that are launched within DigitalOcean.

The builder does _not_ manage images. Once it creates an image, it is up to
you to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

### Required:

* `api_token` (string) - The client TOKEN to use to access your account.
  It can also be specified via environment variable `DIGITALOCEAN_API_TOKEN`, if set.

* `image` (string) - The name (or slug) of the base image to use. This is the
  image that will be used to launch a new droplet and provision it.
  See https://developers.digitalocean.com/documentation/v2/#list-all-images for details on how to get a list of the the accepted image names/slugs.

* `region` (string) - The name (or slug) of the region to launch the droplet in.
  Consequently, this is the region where the snapshot will be available.
  See https://developers.digitalocean.com/documentation/v2/#list-all-regions for the accepted region names/slugs.

* `size` (string) - The name (or slug) of the droplet size to use.
  See https://developers.digitalocean.com/documentation/v2/#list-all-sizes for the accepted size names/slugs.

### Optional:

* `droplet_name` (string) - The name assigned to the droplet. DigitalOcean
  sets the hostname of the machine to this value.

* `private_networking` (boolean) - Set to `true` to enable private networking
  for the droplet being created. This defaults to `false`, or not enabled.

* `snapshot_name` (string) - The name of the resulting snapshot that will
  appear in your account. This must be unique.
  To help make this unique, use a function like `timestamp` (see
  [configuration templates](/docs/templates/configuration-templates.html) for more info)

* `ssh_port` (integer) - The port that SSH will be available on. Defaults to port
  22.

* `ssh_timeout` (string) - The time to wait for SSH to become available
  before timing out. The format of this value is a duration such as "5s"
  or "5m". The default SSH timeout is "1m".

* `ssh_username` (string) - The username to use in order to communicate
  over SSH to the running droplet. Default is "root".

* `state_timeout` (string) - The time to wait, as a duration string,
  for a droplet to enter a desired state (such as "active") before
  timing out. The default state timeout is "6m".

* `user_data` (string) - User data to launch with the Droplet.

## Basic Example

Here is a basic example. It is completely valid as soon as you enter your
own access tokens:

```javascript
{
  "type": "digitalocean",
  "api_token": "YOUR API KEY",
  "image": "ubuntu-12-04-x64",
  "region": "nyc2",
  "size": "512mb"
}
```
