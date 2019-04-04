---
description: |
    The yandex Packer builder is able to create images for use with 
    Yandex.Cloud based on existing images.
layout: docs
page_title: 'Yandex Compute - Builders'
sidebar_current: 'docs-builders-yandex'
---

# Yandex Compute Builder

Type: `yandex`

The `yandex` Packer builder is able to create
[images](https://cloud.yandex.com/docs/compute/concepts/images) for use with
[Yandex Compute Cloud](https://cloud.yandex.com/docs/compute/)
based on existing images.

## Authentication

Yandex.Cloud services authentication requires one of the following security credentials:

-   OAuth token
-   File with Service Account Key
 

### Authentication Using Token

To authenticate with an OAuth token only `token` config key is needed.
Or use the `YC_TOKEN` environment variable.


### Authentication Using Service Account Key File

To authenticate with a service account credential, only `service_account_key_file` is needed.
Or use the `YC_SERVICE_ACCOUNT_KEY_FILE` environment variable.


## Basic Example

``` json
{
  "type": "yandex",
  "token": "YOUR OAUTH TOKEN",
  "folder_id": "YOUR FOLDER ID",
  "source_image_family": "ubuntu-1804-lts",
  "ssh_username": "ubuntu",
  "use_ipv4_nat": "true"
}
```

## Configuration Reference

Configuration options are organized below into two categories: required and
optional. Within each category, the available options are alphabetized and
described.

In addition to the options listed here, a [communicator](/docs/templates/communicator.html) 
can be configured for this builder.

### Required:

-   `folder_id` (string) - The folder ID that will be used to launch instances and store images.
    Alternatively you may set value by environment variable `YC_FOLDER_ID`.
    
-   `token` (string) - OAuth token to use to authenticate to Yandex.Cloud. Alternatively you may set
    value by environment variable `YC_TOKEN`.

-   `source_image_family` (string) - The source image family to create the new image
    from. You can also specify `source_image_id` instead. Just one of a `source_image_id` or 
    `source_image_family` must be specified. Example: `ubuntu-1804-lts`

### Optional:

-   `endpoint` (string) - Non standard api endpoint URL.

-   `instance_cores` (number) - The number of cores available to the instance. 

-   `instance_mem_gb`  (number) - The amount of memory available to the instance, specified in gigabytes.

-   `disk_name` (string) - The name of the disk, if unset the instance name
    will be used.
    
-   `disk_size_gb` (number) - The size of the disk in GB. This defaults to `10`, which is 10GB.

-   `disk_type`  (string) - Specify disk type for the launched instance. Defaults to `network-hdd`.

-   `image_description` (string) - The description of the resulting image.

-   `image_family` (string) -  The family name of the resulting image.

-   `image_labels` (object of key/value strings) - Key/value pair labels to
    apply to the created image.
    
-   `image_name` (string) - The unique name of the resulting image. Defaults to
    `packer-{{timestamp}}`.    

-   `image_product_ids` (list) - License IDs that indicate which licenses are attached to resulting image.

-   `instance_name`  (string) - The name assigned to the instance.
                
-   `labels` (object of key/value strings) - Key/value pair labels to apply to
    the launched instance.
    
-   `machine_type` (string) - The type of virtual machine to launch. This defaults to 'standard-v1'.

-   `metadata` (object of key/value strings) - Metadata applied to the launched
    instance.

-   `serial_log_file` (string) - File path to save serial port output of the launched instance.

-   `service_account_key_file` (string) - Path to file with Service Account key in json format. This 
    is an alternative method to authenticate to Yandex.Cloud. Alternatively you may set environment variable
    `YC_SERVICE_ACCOUNT_KEY_FILE`.

-   `source_image_folder_id` (string) - The ID of the folder containing the source image.

-   `source_image_id` (string) - The source image ID to use to create the new image
    from.

-   `source_image_family` (string) - The source image family to create
    the new image from. The image family always returns its latest image that
    is not deprecated. Example: `ubuntu-1804-lts`.

-   `subnet_id` (string) - The Yandex VPC subnet id to use for 
    the launched instance. Note, the zone of the subnet must match the
    `zone` in which the VM is launched. 

-   `use_internal_ip` (boolean) - If true, use the instance's internal IP address
    instead of its external IP during building.

-   `use_ipv4_nat` (boolean) - If set to `true`, then launched instance will have external internet 
    access. 
 
-   `use_ipv6` (boolean) - Set to `true` to enable IPv6 for the instance being
    created. This defaults to `false`, or not enabled.
-&gt; **Note:** ~> Usage of IPv6 will be available in the future.

-   `state_timeout` (string) - The time to wait for instance state changes.
    Defaults to `5m`.

-   `zone` (string) - The name of the zone to launch the instance.  This defaults to `ru-central1-a`.
