---
description: |
    The vultr Packer builder is able to create new images for use with
    Vultr. The builder takes a source image, runs any provisioning necessary
    on the image after launching it, then snapshots it into a reusable image. This
    reusable image can then be used as the foundation of new servers that are
    launched within Vultr.
layout: docs
page_title: 'Vultr - Builders'
sidebar_current: 'docs-builders-vultr'
---

# Vultr Builder

Type: `vultr`

The `vultr` Packer builder is able to create new images for use with
[Vultr](https://www.vultr.com). The builder takes a source image,
runs any provisioning necessary on the image after launching it, then snapshots
it into a reusable image. This reusable image can then be used as the
foundation of new servers that are launched within Vultr.

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

-   `api_key` (string) - The Vultr API Key to use to access your account.

-   `os_id` (int) - The id of the os to use. This will be the OS that will be used to launch a new instance and provision it. See <a href="https://www.vultr.com/api/#os_os_lists" class="uri">https://www.vultr.com/api/#os_os_list</a>.

-   `region_id` (int) - The id of the region to launch the instance in. See
    <a href="https://www.vultr.com/api/#regions_region_availability" class="uri">https://www.vultr.com/api/#regions_region_availability</a>
    
-   `plan_id` (int) - The id of the plan you wish to use. See
    <a href="https://www.vultr.com/api/#plans_plan_list" class="uri">https://www.vultr.com/api/#plans_plan_list</a>

### Optional:

-   `snapshot_description` (string) - Description of the snapshot.

-   `snapshot_id` (string) -   If you've selected the 'snapshot' (OS 164) operating system, this should be the SNAPSHOTID.

-   `iso_id` (int) - If you've selected the 'custom' (OS 159) operating system, this is the ID of a specific ISO to mount during the deployment.

-   `app_id` (string) - If launching an application (OSID 186), this is the APPID to launch.

-   `enable_ipv6` (boolean) - IPv6 subnet will be assigned to the machine.

-   `enable_private_network` (bool) - Enables private networking support to the new server.

-   `script_id` (string) - If you've not selected a 'custom' (OS 159) operating system, this can be the SCRIPTID of a startup script to execute on boot. 

-   `ssh_key_ids` (array of string) - List of SSH keys to apply to this server on install. Separate keys with commas.

-   `instance_label` (string) - This is a text label that will be shown in the control panel.

-   `userdata` (string) - Base64 encoded user-data.

-   `hostname` (string) - Hostname to assign to this server.

-   `tag` (string) - The tag to assign to this server.

## Basic Example

Here is a Vultr builder example. The vultr_api_key should be replaced with an actual Vultr API Key

``` json
    "variables": {
        "vultr_api_key": "{{ env `VULTR_API_KEY` }}"
    },
    "builders": [{
        "type": "vultr",
        "api_key": "{{ user `vultr_api_key` }}",
        "snapshot_description": "Packer-test-with updates",
        "region_id": 4,
        "plan_id": 402,
        "os_id": 127,
        "ssh_username": "root",
        "stateTimeout": 100
    }],
```
