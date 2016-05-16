---
description: |
    The `lxc` Packer builder builds containers for lxc1. The builder starts an LXC
    container, runs provisioners within this container, then exports the container
    as a tar.gz of the root file system.
layout: docs
page_title: LXC Builder
...

# LXC Builder

Type: `lxc`

The `lxc` Packer builder builds containers for lxc1. The builder starts an LXC
container, runs provisioners within this container, then exports the container
as a tar.gz of the root file system.

The LXC builder requires a modern linux kernel and the `lxc` or `lxc1` package.
This builder does not work with LXD.

## Basic Example

Below is a fully functioning example. 

``` {.javascript}
{
  "builders": [
    {
      "type": "lxc",
      "name": "lxc-trusty",
      "config_file": "/tmp/lxc/config",
      "template_name": "ubuntu",
      "template_environment_vars": [
        "SUITE=trusty"
      ]
    },
    {
      "type": "lxc",
      "name": "lxc-xenial",
      "config_file": "/tmp/lxc/config",
      "template_name": "ubuntu",
      "template_environment_vars": [
        "SUITE=xenial"
      ]
    },
    {
      "type": "lxc",
      "name": "lxc-jessie",
      "config_file": "/tmp/lxc/config",
      "template_name": "debian",
      "template_environment_vars": [
        "SUITE=jessie"
      ]
    }
  ]
}
```

## Configuration Reference

### Required:

-  `config_file` (string) - The path to the lxc configuration file.

-  `template_name` (string) - The LXC template name to use.

-  `template_environment_vars` (array of strings) - Environmental variables to use to build the template with.

### Optional:

-  `target_runlevel` (int) - The minimum run level to wait for the container to reach. Note some distributions (Ubuntu) simulate run levels and may report 5 rather than 3.

-  `output_directory` (string) - The directory in which to save the exported tar.gz.

-  `container_name` (string) - The name of the LXC container. Usually `/var/lib/lxc/containers/<container_name>`.

-  `command_wrapper` (string) - Allows you to specify a wrapper command, such as `ssh` so you can execute packer builds on a remote host.

-  `init_timeout` (string) - The timeout in seconds to wait for the the container to start.

-  `template_parameters` (array of strings) - Environmental variables for running `lxc-create`.

