---
description: |
    The `ucloud-uhost` Packer builder plugin provides the capability to build
    customized images based on an existing base images.
layout: docs
page_title: UCloud Image Builder
sidebar_current: 'docs-builders-ucloud-uhost'
---

# UCloud Image Builder

Type: `ucloud-uhost`

The `ucloud-uhost` Packer builder plugin provides the capability to build
customized images based on an existing base images.

This builder builds an UCloud image by launching an UHost instance from a source image,
provisioning that running machine, and then creating an image from that machine.

\~&gt; **Note:**  This builder only support ssh authenticating with username and given password.

## Configuration Reference

The following configuration options available for building UCloud images. They are
segmented below into two categories: required and optional parameters.

In addition to the options listed here, a
[communicator](../templates/communicator.html) can be configured for this
builder.

### Required:

-   `public_key` - (string) This is the UCloud public key. It must be provided, but it can also be sourced from the `UCLOUD_PUBLIC_KEY` environment variable.

-   `private_key` - (string) This is the UCloud private key. It must be provided, but it can also be sourced from the `UCLOUD_PRIVATE_KEY` environment variable.
  
-   `project_id` - (string) This is the UCloud project id. It must be provided, but it can also be sourced from the `UCLOUD_PROJECT_ID` environment variables.

-   `region` - (string) This is the UCloud region. It must be provided, but it can also be sourced from the `UCLOUD_REGION` environment variables.
  
-   `availability_zone` - (string) This is the UCloud availability zone where UHost instance is located. such as: `cn-bj2-02`. You may refer to [list of availability_zone](https://docs.ucloud.cn/api/summary/regionlist)

-   `instance_type` - (string) The type of UHost instance. You may refer to [list of instance type](https://docs.ucloud.cn/compute/terraform/specification/instance)

-   `image_name` - (string) The name of the user-defined image, which contains 1-63 characters and only support Chinese, English, numbers, '-_,.:[]'.

-   `source_image_id` (string) - This is the ID of base image which you want to create your customized images with.

### Optional:

-   `use_ssh_private_ip` - (boolean) - If this value is true, packer will connect to the created UHost instance via a private ip instead of allocating an EIP (elastic public ip).(Default: `false`).

\~&gt; **Note:**  By default (`use_ssh_private_ip` is `false`), the launched uhost instance will be connecting with extranet by bounding with an EIP  (elastic public ip) automatically, which bandwidth is 30 Mb by default and paid by traffic.

-   `internet_bandwidth` - (string) Maximum bandwidth to the EIP (elastic public ip), measured in Mbps (Mega bit per second). 
    The ranges for bandwidth are: 1-200 to pay by traffic, 1-800 to pay by bandwith. (Default: `1`).
    
-   `internet_charge_mode` -(Optional) The EIP (elastic public ip) charge mode associated to UHost instance. Possible values are: `traffic` as pay by traffic, `bandwidth` as pay by bandwidth. (Default: `traffic`).

-   `vpc_id` - (string) The ID of VPC linked to the UHost instance. If not defined `vpc_id`, the instance will use the default VPC in the current region.

-   `subnet_id` - (string) The ID of subnet under VPC. If  `vpc_id` is defined, the `subnet_id` is mandatory required. If `vpc_id` and `subnet_id` are not defined, the instance will use the default subnet in the current region.

-   `security_group` - (string) The ID of the fire wall associated to UHost instance. If `security_group` is not defined, 
    the instance will use the non-recommended web fire wall, and open port include 22, 3389 by default. It is supported by ICMP fire wall protocols.
    You may refer to [security group](https://docs.ucloud.cn/network/firewall/firewall.html).

-   `image_description` (string) - The description of the image.

-   `instance_name` (string) -  The name of instance, which contains 1-63 characters and only support Chinese, English, numbers, '-', '_', '.'.

-   `boot_disk_type` - (string) The type of boot disk associated to UHost instance. 
    Possible values are: `cloud_ssd` for cloud boot disk, `local_normal` and `local_ssd` for local boot disk. (Default: `cloud_ssd`).
    The  `cloud_ssd` and `local_ssd` are not fully supported by all regions as boot disk type, please proceed to UCloud console for more details.

\~&gt; **Note:** It takes around 10 mins for boot disk initialization when `boot_disk_type` is `local_normal` or `local_ssd`.

-   `image_copy_to_mappings` (array of copied image mappings) - The array of mappings regarding the copied images to the destination regions and projects.
    -   `project_id` (string) - The destination project id, where copying image in.

    -   `region` (string) -  The destination region, where copying image in.

    -   `name` (string) - The copied image name. If not defined, builder will use `image_name` as default name.  

    -   `description` (number) - The copied image description.

## Basic Example

Here is a basic example for build UCloud image.

``` json
{
  "variables": {
    "ucloud_public_key": "{{env `UCLOUD_PUBLIC_KEY`}}",
    "ucloud_private_key": "{{env `UCLOUD_PRIVATE_KEY`}}"
  },
  "builders": [{
    "type":"ucloud-uhost",
    "public_key":"{{user `ucloud_public_key`}}",
    "private_key":"{{user `ucloud_private_key`}}",
    "region":"cn-bj2",
    "image_name":"packer_test_{{timestamp}}",
    "source_image":"uimage-u3d50m",
    "ssh_username":"root",
    "instance_type":"n-basic-2",
  }]
}
```

-&gt; **Note:** Packer can also read the public key and private key from
environmental variables. See the configuration reference in the section above
for more information on what environmental variables Packer will look for.

\~&gt; **Note:** Source image may be deprecated after a while, you can use the tools like `UCloud CLI` to run `ucloud image list` to find one that exists.
