---
description: |
    The `profitbricks` Packer builder is able to create new images for use with
    ProfitBricks. The builder takes a source image, runs any provisioning necessary
    on the image after launching it, then snapshots it into a reusable image. This
    reusable image can then be used as the foundation of new servers that are
    launched within ProfitBricks.
layout: docs
page_title: ProfitBricks Builder
...

# ProfitBricks Builder

Type: `profitbricks`

The `profitbricks` Packer builder is able to create new images for use with
[ProfitBricks](https://www.profitbricks.com). The builder takes a source image,
runs any provisioning necessary on the image after launching it, then snapshots
it into a reusable image. This reusable image can then be used as the foundation
of new servers that are launched within ProfitBricks.

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

-   `pbpasswrod` (string) - ProfitBricks password. It
    can also be specified via environment variable `PROFITBRICKS_PASSWORD`,
    if set.
    
-   `pbusername` (string) - ProfitBricks username. It
    can also be specified via environment variable `PROFITBRICKS_USERNAME`,
    if set.

-   `servername` (string) - The name of the server that will be created.

### Optional:

-   `cores` (int) - Number of server cores default value 4.

-   `disksize` (string) - Desired disk size default value 50gb

-   `disktype` (string) - Desired disk type default value "HDD"

-   `image` (string) - ProfitBricks volume image default value `Ubuntu-16.04`

-   `pburl` (string) - ProfitBricks REST Url.

-   `ram` (int) - RAM size for the server default value 2048.

-   `region` (string) - ProfitBricks region default value "us/las".

## Basic Example

Here is a basic example. It is completely valid as soon as you enter your own
access tokens:

``` {.javascript}
{
  "builders": [
    {
      "type": "profitbricks",
      "image": "Ubuntu-16.04",
      "pbusername": "pb_username",
      "pbpassword": "pb_password",
      "servername": "packer"
    }
}
```
