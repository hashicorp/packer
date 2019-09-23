---
description: |
    The `zstack` Packer builder helps you to build zstack images from an existed base image
layout: docs
page_title: 'Zstack Builder'
sidebar_current: 'docs-builders-zstack'
---

# ZStack Builder

Type: `zstack`

The `zstack` Packer builder helps you to build zstack images from an existed base image

## Configuration Reference

In order to build a ZStack vminstance image, full-fill your configuration file. Necessary attributes
are given below:

### Required-Parameters:

- `type` (string) - This parameter tells which cloud-service-provider you are using, in our case, use 'zstack-vminstance'
- `base_url` (string) - API endpoint for zstack management node, such as `http://192.168.0.1:8080`.
- `access_key` (string) - Your ZStack access key.
- `key_secret` (string) - Your ZStack secret key.
- `image_uuid` (string) - Your ZStack image uuid, like `0d91658fa3184ed8bc0fd85392d1e809`
- `zone_uuid` (string) - Your ZStack zone uuid, like `0d91658fa3184ed8bc0fd85392d1e809`
- `l3network_uuid` (string) - Your ZStack l3network uuid, like `0d91658fa3184ed8bc0fd85392d1e809`
- `instance_offering` (string) - Your ZStack instance_offering uuid, like `0d91658fa3184ed8bc0fd85392d1e809`

### Optional-Parameters

- `show_secret ` (bool) - If it is `false`, then the `key_secret` will be hidden everywhere, default is `false`
- `ssh_username` (string) - ssh username
- `ssh_password` (string) - ssh password of the `username`
- `ssh_public_key_file ` (string) - public key to ssh the target vminstance
- `ssh_private_key_file ` (string) - private key to ssh the target vminstance
- `skip_delete_vminstance  ` (bool) - If it is `true`, then the vm instance will be reserved (usually it is used for debug), default is `false`
- `skip_provision_mod  ` (bool) - if it is `true`, then will skip the provision and ssh steps, default is `false`
- `skip_packer_systemtag  ` (bool) - zstack-packer will add `packer` systemtags on all the resources created by itself, but it only        supported on zstack verison greater or equal to `3.7.0`, so if you use zstack less than `3.7.0`, you must set it to `true`.  default is `false`
- `export_image  ` (bool) - export the final image with a given url, default is `false`
- `image_name ` (string) - the image name which will be created, default is `packer-{{timestamp}}`
- `image_description ` (string) - the image description which will be created, default is empty
- `instance_name ` (string) - the vminstance name which will be created, default is `packer-{{uuid}}`
- `user_data` (string) - the userdata will be inject to vm, like `hello zstack`, default is empty
- `user_data_file` (string) - the userdata file will be inject to vm, like `/tmp/user_data`, default is empty
- `datavolume_image_uuid` (string) - the datavolume base image which will be used to create a data volume, like `0d91658fa3184ed8bc0fd85392d1e809`, default is empty
- `datavolume_size` (string) - the datavolume size which will be used to create a data volume, like `1g`, support 'k', 'm', 'g', 't', 'p'. It cannot appear with `datavolume_image_uuid` meanwhile. default is empty
- `create_with_root` (bool) - you can create root-volume image and data-volume image both if it is `true`. default is `false`
- `mount_path` (string) - the mount path for data volume like `/zstack_builder`, it will be mount to `/dev/vdb`. default is `/builder`
- `filesystem` (string) - the filesystem which will be used to format datavolume like `ext4`, default is `xfs`
- `state_timeout` (string) - timeout wait for `image` or `vminstance` status ready, default is `120s`
- `create_timeout` (string) - timeout wait for zstack async api call polling, default is `60s`
- `poll_replace_str` ([]string) - a simple proxy, e.g: `["10.0.0.5","172.20.0.5"]` means you can use `172.20.0.5` instead of `10.0.0.5` to poll the zstack async api call


## Examples

Here is a basic example for ZStack.

``` json
{
  "builders": [
	  {
		  "type": "zstack-vminstance",
		  "access_key": "<your access key>",
		  "key_secret": "<your secret key>",
		  "ssh_username": "root",
		  "ssh_password": "password",
		  "base_url": "<your zstack url>",
		  "zone_uuid": "<your zone uuid>",
		  "image_uuid": "<your image uuid>",
		  "l3network_uuid": "<your l3network uuid>",
		  "instance_offering": "<your instance-offering uuid>"
	  }
  ],

  "provisioners": [
	  {
		"type": "shell",
		"inline": [
		   "sleep 3",
		   "dd if=/dev/urandom of=/root/test_dd bs=1M count=2"
		]
	 },
	 {
		 "type":"file",
		 "source":"/root/test_dd",
		 "destination":"/tmp/test_dd",
		 "direction":"download"
	 }
 ]
}


```

[Find more examples](https://github.com/hashicorp/packer/tree/master/examples/zstack)
