---
layout: "docs"
page_title: "OpenStack Builder"
description: |-
  The `openstack` Packer builder is able to create new images for use with OpenStack. The builder takes a source image, runs any provisioning necessary on the image after launching it, then creates a new reusable image. This reusable image can then be used as the foundation of new servers that are launched within OpenStack. The builder will create temporary keypairs that provide temporary access to the server while the image is being created. This simplifies configuration quite a bit.
---

# OpenStack Builder

Type: `openstack`

The `openstack` Packer builder is able to create new images for use with
[OpenStack](http://www.openstack.org). The builder takes a source
image, runs any provisioning necessary on the image after launching it,
then creates a new reusable image. This reusable image can then be
used as the foundation of new servers that are launched within OpenStack.
The builder will create temporary keypairs that provide temporary access to
the server while the image is being created. This simplifies configuration
quite a bit.

The builder does _not_ manage images. Once it creates an image, it is up to
you to use it or delete it.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

### Required:

* `flavor` (string) - The ID or full URL for the desired flavor for the
  server to be created.

* `image_name` (string) - The name of the resulting image.

* `password` (string) - The password used to connect to the OpenStack service.
  If not specified, Packer will use the environment variables
  `SDK_PASSWORD` or `OS_PASSWORD` (in that order), if set.

* `source_image` (string) - The ID or full URL to the base image to use.
  This is the image that will be used to launch a new server and provision it.

* `username` (string) - The username used to connect to the OpenStack service.
  If not specified, Packer will use the environment variables
  `SDK_USERNAME` or `OS_USERNAME` (in that order), if set.

### Optional:

* `api_key` (string) - The API key used to access OpenStack. Some OpenStack
  installations require this.
  If not specified, Packer will use the environment variables
  `SDK_API_KEY`, if set.

* `floating_ip` (string) - A specific floating IP to assign to this instance.
  `use_floating_ip` must also be set to true for this to have an affect.

* `floating_ip_pool` (string) - The name of the floating IP pool to use
  to allocate a floating IP. `use_floating_ip` must also be set to true
  for this to have an affect.

* `insecure` (boolean) - Whether or not the connection to OpenStack can be done
  over an insecure connection. By default this is false.

* `networks` (array of strings) - A list of networks by UUID to attach
  to this instance.

* `openstack_provider` (string) - A name of a provider that has a slightly
  different API model. Currently supported values are "openstack" (default),
  and "rackspace".

* `project` (string) - The project name to boot the instance into. Some
  OpenStack installations require this.
  If not specified, Packer will use the environment variables
  `SDK_PROJECT` or `OS_TENANT_NAME` (in that order), if set.

* `provider` (string) - The provider used to connect to the OpenStack service.
  If not specified, Packer will use the environment variables `SDK_PROVIDER` 
  or `OS_AUTH_URL` (in that order), if set. 
  For Rackspace this should be `rackspace-us` or `rackspace-uk`.

* `proxy_url` (string)

* `security_groups` (array of strings) - A list of security groups by name
  to add to this instance.

* `region` (string) - The name of the region, such as "DFW", in which
  to launch the server to create the AMI.
  If not specified, Packer will use the environment variables
  `SDK_REGION` or `OS_REGION_NAME` (in that order), if set.
  For a `provider` of "rackspace", it is required to specify a region,
  either using this option or with an environment variable. For other
  providers, including a private cloud, specifying a region is optional.

* `ssh_port` (integer) - The port that SSH will be available on. Defaults to port
  22.

* `ssh_timeout` (string) - The time to wait for SSH to become available
  before timing out. The format of this value is a duration such as "5s"
  or "1m". The default SSH timeout is "5m".

* `ssh_username` (string) - The username to use in order to communicate
  over SSH to the running server. The default is "root".

* `ssh_interface` (string) - The type of interface to connect via SSH. Values
  useful for Rackspace are "public" or "private", and the default behavior is
  to connect via whichever is returned first from the OpenStack API.

* `tenant_id` (string) - Tenant ID for accessing OpenStack if your
  installation requires this.

* `use_floating_ip` (boolean) - Whether or not to use a floating IP for
  the instance. Defaults to false.

* `rackconnect_wait` (boolean) - For rackspace, whether or not to wait for
  Rackconnect to assign the machine an IP address before connecting via SSH.
  Defaults to false.

## Basic Example: Rackspace public cloud

Here is a basic example. This is a working example to build a
Ubuntu 12.04 LTS (Precise Pangolin) on Rackspace OpenStack cloud offering.

```javascript
{
  "type": "openstack",
  "username": "",
  "api_key": "",
  "openstack_provider": "rackspace",
  "provider": "rackspace-us",
  "region": "DFW",
  "ssh_username": "root",
  "image_name": "Test image",
  "source_image": "23b564c9-c3e6-49f9-bc68-86c7a9ab5018",
  "flavor": "2"
}
```

## Basic Example: Private OpenStack cloud

This example builds an Ubuntu 14.04 image on a private OpenStack cloud,
powered by Metacloud.

```javascript
{
  "type": "openstack",
  "ssh_username": "root",
  "image_name": "ubuntu1404_packer_test_1",
  "source_image": "91d9c168-d1e5-49ca-a775-3bfdbb6c97f1",
  "flavor": "2"
}
```

In this case, the connection information for connecting to OpenStack
doesn't appear in the template. That is because I source a standard
OpenStack script with environment variables set before I run this. This
script is setting environment variables like:

* `OS_AUTH_URL`
* `OS_TENANT_ID`
* `OS_USERNAME`
* `OS_PASSWORD`

## Troubleshooting

*I get the error "Missing or incorrect provider"*

* Verify your "username", "password" and "provider" settings.

*I get the error "Missing endpoint, or insufficient privileges to access endpoint"*

* Verify your "region" setting.
