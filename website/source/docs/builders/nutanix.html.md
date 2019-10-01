---
description: |
    The nutanix Packer builder builds Nutanix images from Virtual Machines. The builder starts
    a Nutanix VM, runs provisioners within this container, then saves the the image within Nutanix.
layout: docs
page_title: 'Nutanix - Builders'
sidebar_current: 'docs-builders-nutanix'
---

# Nutanix Builder

Type: `nutanix`

The `nutanix` Packer builder builds [Nutanix](https://www.nutanix.com/) images using
Nutanix. The builder starts a Nutanix VM, runs provisioners against this
VM, then copies the disk for reuse.


The Nutanix builder relies on the [Nutanix Intentful API](https://www.nutanix.dev/reference/prism_central/v3) to build VMs using Nutanix resources.  This allows you to build VMs with temporary compute and memory within your infrastructure directly from your laptop or CICD environment, without all of the hardware requirements local to the process running the build.

## Basic Example

Below is an example template to build.  Similar environment variables supported by the [Nutanix Terraform Project](https://www.terraform.io/docs/providers/nutanix/#environment-variables) can be used for talking to your cluster. 


> _Note - Currently proxy configurations are not supported at this time._

```json
{
  "type": "nutanix",
  "new_image_name": "linux2-x86_64-{{user `timestamp`}}",
  "nutanix_endpoint": "https://198.162.88.1:9440",
  "nutanix_insecure": true,
  "nutanix_username": "{{user `nutanix_user`}}",
  "nutanix_password": "{{user `nutanix_password`}}",
  "winrm_username": "Administrator",
  "winrm_password": "{{user `elevated_password`}}",
  "winrm_use_ssl": true,
  "communicator": "winrm",
  "winrm_timeout": "30m",
  "shutdown_command": "C:\\Windows\\system32\\sysprep\\sysprep.exe /generalize /oobe /unattend:e:\\answer_files\\deployautounattend.xml /quiet /shutdown",
  "shutdown_timeout": "15m",
  "pause_before_connecting": "1m",
  "metadata": {
      "kind": "vm"
  },
  "spec": {
    "resources": {
      "num_threads_per_core": 2,
      "num_vcpus_per_socket": 2,
      "memory_size_mib": 8192,
      "nic_list": [{
          "subnet_reference": {
            "kind": "subnet",
            "uuid": "a000000-a000-0000-0000-a00000000"
          }
      }],
      "boot_config": {
        "boot_device_order_list": [
          "CDROM",
          "DISK"
        ]
      },
      "disk_list": [
        {
          "data_source_reference": {
            "kind": "image",
            "uuid": "a000000-a000-0000-0000-a00000000"
          },
          "device_properties": {
            "disk_address": {
              "adapter_type": "IDE",
              "device_index": 0
            },
            "device_type": "CDROM"
          }
        },
        {
          "data_source_reference": {
            "kind": "image",
            "uuid": "a000000-a000-0000-0000-a00000000"
          },
          "device_properties": {
            "disk_address": {
              "adapter_type": "IDE",
              "device_index": 1
            },
            "device_type": "CDROM"
          }
        }
      ]
    }
  }
}
```

## Configuration Reference

Configuration options are organized below into two categories: required and
optional. Within each category, the available options are alphabetized and
described.

### Required:

You must specify (only) one of `commit`, `discard`, or `export_path`.

-  `nutanix_endpoint` (string) - Host endpoint. No protocol necessary. Will load from `NUTANIX_ENDPOINT` environment variable as alternative if empty.
-  `nutanix_username` (string) - The cluster username to connect with. Will load from `NUTANIX_USERNAME` environment variable as alternative if empty.
-  `nutanix_password` (string) - The cluster password to connect with. Will load from `NUTANIX_PASSWORD` environment variable as alternative if empty.
-  `new_image_name` (string) - The image name to assign the saved image upon completion.

### Optional

-  `nutanix_insecure` (bool) - Whether to check the Nutanix endpoint certificate. Default `false`.  Will load from `NUTANIX_INSECURE` environment variable as alternative if empty.

## Provisioner Authentication

The Nutanix builder will generate and assign a random ssh key to the running VM to authenticate and provision the image.  If cloud-init values are already being overridden, then it is assumed the user is passing in the authentication information for the ssh communicator.

Winrm will create a basic sysprep config file if no authentication information is provided. The sysprep file will be mounted on a drive for the image to pick-up assuming it is already prepped.  Alternatively, if the username and password are known on the Windows machine, they can be passed in as the winrm communicator credentials at runtime.
