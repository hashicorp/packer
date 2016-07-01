---
description: |
    The `lxd` Packer builder builds containers for LXD. The builder starts an LXD
    container, runs provisioners within this container, then saves the container
    as an LXD image.
layout: docs
page_title: LXD Builder
...

# LXD Builder

Type: `lxd`

The `lxd` Packer builder builds containers for LXD. The builder starts an LXD
container, runs provisioners within this container, then saves the container
as an LXD image.

The LXD builder requires a modern linux kernel and the `lxd` package.
This builder does not work with LXC.

## Basic Example

Below is a fully functioning example. 

``` {.javascript}
{
  "builders": [
    {
      "type": "lxd",
      "name": "lxd-xenial",
      "image": "ubuntu-daily:xenial",
      "output_image": "ubuntu-xenial"
    }
  ]
}
```

## Configuration Reference

### Required:

-  `image` (string) - The source image to use when creating the build container. This can be a (local or remote) image (name or fingerprint). E.G. my-base-image, ubuntu-daily:x, 08fababf6f27...
    Note: The builder may appear to pause if required to download a remote image, as they are usually 100-200MB. `/var/log/lxd/lxd.log` will mention starting such downloads.

### Optional:

-  `name` (string) - The name of the started container. Defaults to `packer-$PACKER_BUILD_NAME`.

-  `output_image` (string) - The name of the output artifact. Defaults to `name`

