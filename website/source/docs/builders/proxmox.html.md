---
description: |
    The proxmox Packer builder is able to create new images for use with
    Proxmox VE. The builder takes an ISO source, runs any provisioning
    necessary on the image after launching it, then creates a virtual machine
    template.
layout: docs
page_title: 'Proxmox - Builders'
sidebar_current: 'docs-builders-proxmox'
---

# Proxmox Builder

Type: `proxmox`

The `proxmox` Packer builder is able to create new images for use with
[Proxmox](https://www.proxmox.com/en/proxmox-ve). The builder takes an ISO
image, runs any provisioning necessary on the image after launching it, then
creates a virtual machine template. This template can then be used as to
create new virtual machines within Proxmox.

The builder does *not* manage templates. Once it creates a template, it is up
to you to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `proxmox_url` (string) - URL to the Proxmox API, including the full path,
    so `https://<server>:<port>/api2/json` for example.
    Can also be set via the `PROXMOX_URL` environment variable.

-   `username` (string) - Username when authenticating to Proxmox, including
    the realm. For example `user@pve` to use the local Proxmox realm.
    Can also be set via the `PROXMOX_USERNAME` environment variable.

-   `password` (string) - Password for the user.
    Can also be set via the `PROXMOX_PASSWORD` environment variable.

-   `node` (string) - Which node in the Proxmox cluster to start the virtual
    machine on during creation.

-   `iso_file` (string) - Path to the ISO file to boot from, expressed as a
    proxmox datastore path, for example
    `local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso`

### Optional:
-   `insecure_skip_tls_verify` (bool) - Skip validating the certificate.

-   `vm_name` (string) - Name of the virtual machine during creation. If not
    given, a random uuid will be used.

-   `vm_id` (int) - The ID used to reference the virtual machine. This will
    also be the ID of the final template. If not given, the next free ID on
    the node will be used.

-   `memory` (int) - How much memory, in megabytes, to give the virtual
    machine. Defaults to `512`.

-   `cores` (int) - How many CPU cores to give the virtual machine. Defaults
    to `1`.

-   `sockets` (int) - How many CPU sockets to give the virtual machine.
    Defaults to `1`

-   `os` (string) - The operating system. Can be `wxp`, `w2k`, `w2k3`, `w2k8`,
    `wvista`, `win7`, `win8`, `win10`, `l24` (Linux 2.4), `l26` (Linux 2.6+),
    `solaris` or `other`. Defaults to `other`.

-   `network_adapters` (array of objects) - Network adapters attached to the
    virtual machine. Example:

    ```json
    [
      {
        "model": "virtio",
        "bridge": "vmbr0",
        "vlan_tag": "10"
      }
    ]
    ```

    -   `bridge` (string) - Required. Which Proxmox bridge to attach the
        adapter to.

    -   `model` (string) - Model of the virtual network adapter. Can be
        `rtl8139`, `ne2k_pci`, `e1000`, `pcnet`, `virtio`, `ne2k_isa`,
        `i82551`, `i82557b`, `i82559er`, `vmxnet3`, `e1000-82540em`,
        `e1000-82544gc` or `e1000-82545em`. Defaults to `e1000`.

    -   `mac_address` (string) - Give the adapter a specific MAC address. If
        not set, defaults to a random MAC.

    -   `vlan_tag` (string) - If the adapter should tag packets. Defaults to
        no tagging.

-   `disks` (array of objects) - Disks attached to the virtual machine.
    Example:

    ```json
    [
      {
        "type": "scsi",
        "disk_size": "5G",
        "storage_pool": "local-lvm",
        "storage_pool_type": "lvm"
      }
    ]
    ```

    -   `storage_pool` (string) - Required. Name of the Proxmox storage pool
        to store the virtual machine disk on. A `local-lvm` pool is allocated
        by the installer, for example.

    -   `storage_pool_type` (string) - Required. The type of the pool, can
        be `lvm`, `lvm-thin`, `zfs` or `directory`.

    -   `type` (string) - The type of disk. Can be `scsi`, `sata`, `virtio` or
        `ide`. Defaults to `scsi`.

    -   `disk_size` (string) - The size of the disk, including a unit suffix, such
        as `10G` to indicate 10 gigabytes.

    -   `cache_mode` (string) - How to cache operations to the disk. Can be
        `none`, `writethrough`, `writeback`, `unsafe` or `directsync`.
        Defaults to `none`.

    -   `format` (string) - The format of the file backing the disk. Can be
        `raw`, `cow`, `qcow`, `qed`, `qcow2`, `vmdk` or `cloop`. Defaults to
        `raw`.

-   `template_name` (string) - Name of the template. Defaults to the generated
    name used during creation.

-   `template_description` (string) - Description of the template, visible in
    the Proxmox interface.

-   `unmount_iso` (bool) - If true, remove the mounted ISO from the template
    after finishing. Defaults to `false`.

-   `qemu_agent` (boolean) - Disables QEMU Agent option for this VM. When enabled,
    then `qemu-guest-agent` must be installed on the guest. When disabled, then 
    `ssh_host` should be used. Defaults to `true`.

## Example: Fedora with kickstart

Here is a basic example creating a Fedora 29 server image with a Kickstart
file served with Packer's HTTP server. Note that the iso file needs to be
manually downloaded.

``` json
{
  "variables": {
    "username": "apiuser@pve",
    "password": "supersecret"
  },
  "builders": [
    {
      "type": "proxmox",
      "proxmox_url": "https://my-proxmox.my-domain:8006/api2/json",
      "insecure_skip_tls_verify": true,
      "username": "{{user `username`}}",
      "password": "{{user `password`}}",

      "node": "my-proxmox",
      "network_adapters": [
        {
          "bridge": "vmbr0"
        }
      ],
      "disks": [
        {
          "type": "scsi",
          "disk_size": "5G",
          "storage_pool": "local-lvm",
          "storage_pool_type": "lvm"
        }
      ],

      "iso_file": "local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso",
      "http_directory":"config",
      "boot_wait": "10s",
      "boot_command": [
        "<up><tab> ip=dhcp inst.cmdline inst.ks=http://{{.HTTPIP}}:{{.HTTPPort}}/ks.cfg<enter>"
      ],

      "ssh_username": "root",
      "ssh_timeout": "15m",
      "ssh_password": "packer",

      "unmount_iso": true,
      "template_name": "fedora-29",
      "template_description": "Fedora 29-1.2, generated on {{ isotime \"2006-01-02T15:04:05Z\" }}"
    }
  ]
}
```
