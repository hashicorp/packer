---
description: |
    The `alicloud-ecs` Packer builder plugin provide the capability to build
    customized images based on an existing base images.
layout: docs
page_title: Alicloud Image Builder
sidebar_current: 'docs-builders-alicloud-ecs'
---

# Alicloud Image Builder

Type: `alicloud-ecs`

The `alicloud-ecs` Packer builder plugin provide the capability to build
customized images based on an existing base images.

## Configuration Reference

The following configuration options are available for building Alicloud images.
In addition to the options listed here,
a [communicator](../templates/communicator.html) can be configured for this
builder.

### Required:

-   `access_key` (string) - This is the Alicloud access key. It must be provided,
    but it can also be sourced from the `ALICLOUD_ACCESS_KEY` environment
    variable.

-   `image_name` (string) - The name of the user-defined image, \[2, 128\] English
    or Chinese characters. It must begin with an uppercase/lowercase letter or
    a Chinese character, and may contain numbers, `_` or `-`. It cannot begin with
    `http://` or `https://`.

-   `instance_type` (string) - Type of the instance. For values, see [Instance
    Type Table](https://www.alibabacloud.com/help/doc-detail/25378.htm?spm=a3c0i.o25499en.a3.9.14a36ac8iYqKRA).
    You can also obtain the latest instance type table by invoking the [Querying
    Instance Type Table](https://intl.aliyun.com/help/doc-detail/25620.htm?spm=a3c0i.o25499en.a3.6.Dr1bik)
    interface.

-   `region` (string) - This is the Alicloud region. It must be provided, but it
    can also be sourced from the `ALICLOUD_REGION` environment variables.

-   `secret_key` (string) - This is the Alicloud secret key. It must be provided,
    but it can also be sourced from the `ALICLOUD_SECRET_KEY` environment
    variable.

-   `source_image` (string) - This is the base image id which you want to create
    your customized images.

### Optional:

-   `force_stop_instance` (boolean) - Whether to force shutdown upon device restart.
    The default value is `false`.

    If it is set to `false`, the system is shut down normally; if it is set to
    `true`, the system is forced to shut down.

-   `disable_stop_instance` (boolean) - If this option is set to `true`, Packer will not stop the instance
    for you, and you need to make sure the instance will be stopped in the final provisioner command. Otherwise,
    Packer will timeout while waiting the instance to be stopped. This option is provided for some specific
    scenarios that you want to stop the instance by yourself. E.g., Sysprep a windows which may shutdown the instance
    within its command. The default value is `false`.     

-   `image_copy_names` (array of string) - The name of the destination image, \[2,
    128\] English or Chinese characters. It must begin with an uppercase/lowercase
    letter or a Chinese character, and may contain numbers, `_` or `-`. It cannot
    begin with `http://` or `https://`.

-   `image_copy_regions` (array of string) - Copy to the destination regionIds.

-   `image_description` (string) - The description of the image, with a length
    limit of 0 to 256 characters. Leaving it blank means null, which is the
    default value. It cannot begin with `http://` or `https://`.

-   `image_disk_mappings` (array of image disk mappings) - Add one or more data disks
    to the image.

    -   `disk_category` (string) - Category of the data disk. Optional values are:
        -   `cloud` - general cloud disk
        -   `cloud_efficiency` - efficiency cloud disk
        -   `cloud_ssd` - cloud SSD

        Default value: cloud.

    -   `disk_delete_with_instance` (boolean) - Whether or not the disk is released along with the instance:
        -   True indicates that when the instance is released, this disk will be released with it
        -   False indicates that when the instance is released, this disk will be retained.

    -   `disk_description` (string) - The value of disk description is blank by default. \[2, 256\] characters. The disk description will appear on the console. It cannot begin with `http://` or `https://`.

    -   `disk_device` (string) - Device information of the related instance: such as
        `/dev/xvdb` It is null unless the Status is In\_use.

    -   `disk_name` (string) - The value of disk name is blank by default. \[2, 128\]
        English or Chinese characters, must begin with an uppercase/lowercase letter
        or Chinese character. Can contain numbers, `.`, `_` and `-`. The disk name
        will appear on the console. It cannot begin with `http://` or `https://`.

    -   `disk_size` (number) - Size of the system disk, in GB, values range:
        -   `cloud` - 5 ~ 2000
        -   `cloud_efficiency` - 20 ~ 2048
        -   `cloud_ssd` - 20 ~ 2048

        The value should be equal to or greater than the size of the specific SnapshotId.

    -   `disk_snapshot_id` (string) - Snapshots are used to create the data disk
        After this parameter is specified, Size is ignored. The actual size of the
        created disk is the size of the specified snapshot.

        Snapshots from on or before July 15, 2013 cannot be used to create a disk.

-   `image_force_delete` (boolean) - If this value is true, when the target image name
    is duplicated with an existing image, it will delete the existing image and
    then create the target image, otherwise, the creation will fail. The default
    value is false.

-   `image_force_delete_snapshots` (boolean) - If this value is true, when delete the
    duplicated existing image, the source snapshot of this image will be delete
    either.

-   `image_share_account` (array of string) - The IDs of to-be-added Aliyun
    accounts to which the image is shared. The number of accounts is 1 to 10. If
    number of accounts is greater than 10, this parameter is ignored.

-   `image_version` (string) - The version number of the image, with a length limit
    of 1 to 40 English characters.

-   `instance_name` (string) - Display name of the instance, which is a string of
    2 to 128 Chinese or English characters. It must begin with an
    uppercase/lowercase letter or a Chinese character and can contain numerals,
    `.`, `_`, or `-`. The instance name is displayed on the Alibaba Cloud
    console. If this parameter is not specified, the default value is InstanceId
    of the instance. It cannot begin with `http://` or `https://`.

-   `internet_charge_type` (string) - Internet charge type, which can be
    `PayByTraffic` or `PayByBandwidth`. Optional values:
    -   `PayByBandwidth`
    -   `PayByTraffic`

    If this parameter is not specified, the default value is `PayByBandwidth`.
    For the regions out of China, currently only support `PayByTraffic`, you must
    set it manfully.

-   `internet_max_bandwidth_out` (string) - Maximum outgoing bandwidth to the public
    network, measured in Mbps (Mega bits per second).

    Value range:
    -   `PayByBandwidth`: \[0, 100\]. If this parameter is not specified, API automatically sets it to 0 Mbps.
    -   `PayByTraffic`: \[1, 100\]. If this parameter is not specified, an error is returned.

-   `io_optimized` (boolean) - Whether an ECS instance is I/O optimized or not.
    The default value is `false`.

-   `security_group_id` (string) - ID of the security group to which a newly
    created instance belongs. Mutual access is allowed between instances in one
    security group. If not specified, the newly created instance will be added to
    the default security group. If the default group doesn’t exist, or the number
    of instances in it has reached the maximum limit, a new security group will
    be created automatically.

-   `security_group_name` (string) - The security group name. The default value is
    blank. \[2, 128\] English or Chinese characters, must begin with an
    uppercase/lowercase letter or Chinese character. Can contain numbers, `.`,
    `_` or `-`. It cannot begin with `http://` or `https://`.

-   `security_token` (string) - STS access token, can be set through template or by exporting
     as environment variable such as `export SecurityToken=value`.

-   `skip_region_validation` (boolean) - The region validation can be skipped if this
    value is true, the default value is false.

-   `temporary_key_pair_name` (string) - The name of the temporary key pair to
    generate. By default, Packer generates a name that looks like `packer_<UUID>`,
    where `<UUID>` is a 36 character unique identifier.

-   `TLSHandshakeTimeout` (int) - When happen "net/http: TLS handshake timeout" problem, set this environment variable
     to a bigger such as `export TLSHandshakeTimeout=30`, it will set the TLS handshake timeout value to 30s.

-   `user_data` (string) - The UserData of an instance must be encoded in `Base64`
    format, and the maximum size of the raw data is `16 KB`.

-   `user_data_file` (string) - The file name of the userdata.

-   `vpc_cidr_block` (string) - Value options: `192.168.0.0/16` and `172.16.0.0/16`.
    When not specified, the default value is `172.16.0.0/16`.

-   `vpc_id` (string) - VPC ID allocated by the system.

-   `vpc_name` (string) - The VPC name. The default value is blank. \[2, 128\]
    English or Chinese characters, must begin with an uppercase/lowercase letter
    or Chinese character. Can contain numbers, `_` and `-`. The disk description
    will appear on the console. Cannot begin with `http://` or `https://`.

-   `vswitch_id` (string) - The ID of the VSwitch to be used.

-   `zone_id` (string) - ID of the zone to which the disk belongs.

-   `ssh_private_ip` (boolean) - If this value is true, packer will connect to the ECS created through private ip
    instead of allocating a public ip or an EIP. The default value is false.
    
-   `tags` (object of key/value strings) - Tags applied to the destination image.

## Basic Example

Here is a basic example for Alicloud.

``` json
{
  "variables": {
    "access_key": "{{env `ALICLOUD_ACCESS_KEY`}}",
    "secret_key": "{{env `ALICLOUD_SECRET_KEY`}}"
  },
  "builders": [{
    "type":"alicloud-ecs",
    "access_key":"{{user `access_key`}}",
    "secret_key":"{{user `secret_key`}}",
    "region":"cn-beijing",
    "image_name":"packer_test2",
    "source_image":"centos_7_04_64_20G_alibase_201701015.vhd",
    "ssh_username":"root",
    "instance_type":"ecs.n1.tiny",
    "io_optimized":"true",
    "image_force_delete":"true"
  }],
  "provisioners": [{
    "type": "shell",
    "inline": [
      "sleep 30",
      "yum install redis.x86_64 -y"
    ]
  }]
}
```

~&gt; Note: Images can become deprecated after a while; run
`aliyun ecs DescribeImages` to find one that exists.

See the
[examples/alicloud](https://github.com/hashicorp/packer/tree/master/examples/alicloud)
folder in the packer project for more examples.
