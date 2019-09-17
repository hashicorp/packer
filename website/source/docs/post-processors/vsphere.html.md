---
description: |
    The Packer vSphere post-processor takes an artifact and uploads it to a vSphere endpoint.
layout: docs
page_title: 'vSphere - Post-Processors'
sidebar_current: 'docs-post-processors-vsphere'
---

# vSphere Post-Processor

Type: `vsphere`

The Packer vSphere post-processor takes an artifact and uploads it to a vSphere endpoint.
The artifact must have a vmx/ova/ovf image.

## Configuration

There are many configuration options available for the post-processor. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

Required:

-   `cluster` (string) - The cluster to upload the VM to.

-   `datacenter` (string) - The name of the datacenter within vSphere to add
    the VM to.

-   `datastore` (string) - The name of the datastore to store this VM. This is
    *not required* if `resource_pool` is specified.

-   `host` (string) - The vSphere host that will be contacted to perform the VM
    upload.

-   `password` (string) - Password to use to authenticate to the vSphere
    endpoint.

-   `username` (string) - The username to use to authenticate to the vSphere
    endpoint.

-   `vm_name` (string) - The name of the VM once it is uploaded.

Optional:

-   `esxi_host` (string) - Target vSphere host. Used to assign specific esx
    host to upload the resulting VM to, when a vCenter Server is used as
    `host`. Can be either a hostname (e.g. "packer-esxi1", requires proper DNS
    setup and/or correct DNS search domain setting) or an ipv4 address.

-   `disk_mode` (string) - Target disk format. See `ovftool` manual for
    available options. By default, "thick" will be used.

-   `insecure` (boolean) - Whether or not the connection to vSphere can be done
    over an insecure connection. By default this is false.

-   `keep_input_artifact` (boolean) - When `true`, preserve the local VM files,
    even after importing them to vsphere. Defaults to `false`.

-   `resource_pool` (string) - The resource pool to upload the VM to.

-   `vm_folder` (string) - The folder within the datastore to store the VM.

-   `vm_network` (string) - The name of the VM network this VM will be added
    to.

-   `overwrite` (boolean) - If it's true force the system to overwrite the
    existing files instead create new ones. Default is false

-   `options` (array of strings) - Custom options to add in ovftool. See
    `ovftool   --help` to list all the options
