---
description: |
    The `lxd` Packer builder builds containers for LXD. The builder starts an LXD
    container, runs provisioners within this container, then saves the container
    as an LXD image.
layout: docs
page_title: 'LXD - Builders'
sidebar_current: 'docs-builders-lxd'
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
      "output_image": "ubuntu-xenial",
      "publish_properties": {
        "description": "Trivial repackage with Packer"
      }
    }
  ]
}
```


## Configuration Reference

### Required:

-  `image` (string) - The source image to use when creating the build
   container. This can be a (local or remote) image (name or fingerprint). E.G.
   `my-base-image`, `ubuntu-daily:x`, `08fababf6f27`, ...

    ~&gt; Note: The builder may appear to pause if required to download
    a remote image, as they are usually 100-200MB. `/var/log/lxd/lxd.log` will
    mention starting such downloads.

### Optional:

-  `init_sleep` (string) - The number of seconds to sleep between launching the
   LXD instance and provisioning it; defaults to 3 seconds.

-  `name` (string) - The name of the started container. Defaults to
   `packer-$PACKER_BUILD_NAME`.

-  `output_image` (string) - The name of the output artifact. Defaults to
   `name`.

-  `command_wrapper` (string) - Lets you prefix all builder commands, such as
   with `ssh` for a remote build host. Defaults to `""`.

-  `publish_properties` (map[string]string) - Pass key values to the publish
   step to be set as properties on the output image. This is most helpful to
   set the description, but can be used to set anything needed.
   See https://stgraber.org/2016/03/30/lxd-2-0-image-management-512/
   for more properties.
   
-  `launch_config` (map[string]string) - List of key/value pairs you wish to
   pass to `lxc launch` via `--config`. Defaults to empty.
