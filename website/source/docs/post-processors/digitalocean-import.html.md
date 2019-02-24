---
description: |
    The Packer DigitalOcean Import post-processor takes an image artifact
    from various builders and imports it to DigitalOcean.
layout: docs
page_title: 'DigitalOcean Import - Post-Processors'
sidebar_current: 'docs-post-processors-digitalocean-import'
---

# DigitalOcean Import Post-Processor

Type: `digitalocean-import`

The Packer DigitalOcean Import post-processor takes an image artifact from
various builders and imports it to DigitalOcean.

## How Does it Work?

The import process operates uploading a temporary copy of the image to
DigitalOcean Spaces and then importing it as a custom image via the
DigialOcean API. The temporary copy in Spaces can be discarded after the
import is complete.

For information about the requirements to use an image for a DigitalOcean
Droplet, see DigitalOcean's [Custom Images documentation](https://www.digitalocean.com/docs/images/custom-images/overview/).

## Configuration

There are some configuration options available for the post-processor.

Required:

-   `api_token` (string) - A personal access token used to communicate with
    the DigitalOcean v2 API. This may also be set using the
    `DIGITALOCEAN_API_TOKEN` environmental variable.

-   `spaces_key` (string) - The access key used to communicate with Spaces.
    This may also be set using the `DIGITALOCEAN_SPACES_ACCESS_KEY`
    environmental variable.

-   `spaces_secret` (string) - The secret key used to communicate with Spaces.
    This may also be set using the `DIGITALOCEAN_SPACES_SECRET_KEY`
    environmental variable.

-   `spaces_region` (string) - The name of the region, such as `nyc3`, in which
    to upload the image to Spaces.

-   `space_name` (string) - The name of the specific Space where the image file
    will be copied to for import. This Space must exist when the
    post-processor is run.

-   `image_name` (string) - The name to be used for the resulting DigitalOcean
    custom image.

-   `image_regions` (array of string) - A list of DigitalOcean regions, such
    as `nyc3`, where the resulting image will be available for use in creating
    Droplets.

Optional:

-   `image_description` (string) - The description to set for the resulting
    imported image.

-   `image_distribution` (string) - The name of the distribution to set for
    the resulting imported image.

-   `image_tags` (array of strings) - A list of tags to apply to the resulting
    imported image.

-   `skip_clean` (boolean) - Whether we should skip removing the image file
    uploaded to Spaces after the import process has completed. "true" means
    that we should leave it in the Space, "false" means to clean it out.
    Defaults to `false`.

-   `space_object_name` (string) - The name of the key used in the Space where
    the image file will be copied to for import. If not specified, this will default to "packer-import-{{timestamp}}".

-   `timeout` (number) - The length of time in minutes to wait for individual
    steps in the process to successfully complete. This includes both importing
    the image from Spaces as well as distributing the resulting image to
    additional regions. If not specified, this will default to 20.

## Basic Example

Here is a basic example:

``` json
{
  "type": "digitalocean-import",
  "api_token": "{{user `token`}}",
  "spaces_key": "{{user `key`}}",
  "spaces_secret": "{{user `secret`}}",
  "spaces_region": "nyc3",
  "space_name": "import-bucket",
  "image_name": "ubuntu-18.10-minimal-amd64",
  "image_description": "Packer import {{timestamp}}",
  "image_regions": ["nyc3", "nyc2"],
  "image_tags": ["custom", "packer"]
}
```
