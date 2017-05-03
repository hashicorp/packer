---
layout: "docs"
page_title: "OpenStack Builder"
description: |-
  The `openstack-id3` Packer builder is able to create new images for use with OpenStack. The builder takes a source image, runs any provisioning necessary on the image after launching it, then creates a new reusable image. This reusable image can then be used as the foundation of new servers that are launched within OpenStack. The builder will create temporary keypairs that provide temporary access to the server while the image is being created. This simplifies configuration quite a bit.
---

# OpenStack Builder

Type: `openstack-id3`

The `openstack-id3` Packer builder is able to create new images for use with
[OpenStack](http://www.openstack.org). It communicates with both identity 
version 2 and 3 and currently use version 2 compute and network APIs 
(What the Gophercloud v1 API supports). The builder takes a source
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

### Authentication

How you authenticate with your provider is first of all based on what identity
version is supported. Some providers might have a more custom authentication
setup. Make sure you know what parameters your provider requires.

### Identity Version 2

* `provider` (string) - The provider used to connect to the OpenStack service
  represented as the URL to the authentication endpoint.
  (Example: https://some.identity.endpoint/v2)
  If not specified, Packer will use the environment variables `SDK_PROVIDER` 
  or `OS_AUTH_URL` (in that order), if set. 
  For Rackspace this should be `rackspace-us` or `rackspace-uk`.

* `username` (string) - The username used to connect to the OpenStack service.
  If not specified, Packer will use the environment variables
  `SDK_USERNAME` or `OS_USERNAME` (in that order), if set.

* `password` (string) - The password used to connect to the OpenStack service.
  If not specified, Packer will use the environment variables
  `SDK_PASSWORD` or `OS_PASSWORD` (in that order), if set.

* `tenant_id` (string) - Tenant ID for accessing OpenStack if your
  installation requires this. Packer will use the environment variable
  `OS_TENANT_ID` or `OS_PROJECT_ID` (in that order) if set.

* `project_id` can also be used and will behave just like `tenant_id`.
  
### Identity Version 3

When authenticating using identity version 3, you need to specify what domain
and what project in that domain you are scoping to. This is required for Packer
to be able to access your resources.

* `provider` (string) - The provider used to connect to the OpenStack service
  represented as the URL to the authentication endpoint.
  (Example: https://some.identity.endpoint/v3)
  If not specified, Packer will use the environment variables `SDK_PROVIDER` 
  or `OS_AUTH_URL` (in that order), if set.

* `username` (string) - The username used to connect to the OpenStack service.
  If not specified, Packer will use the environment variables
  `SDK_USERNAME` or `OS_USERNAME` (in that order), if set.

* `user_id` (string) - The User Id used to connect to the Openstack service.
  If not specified, Packer will use the environment variable `OS_USER_ID`.
  It's fairly common to stick to `username`, but in some instances
  authenticating using with `user_id` can be useful.

* `password` (string) - The password used to connect to the OpenStack service.
  If not specified, Packer will use the environment variables
  `SDK_PASSWORD` or `OS_PASSWORD` (in that order), if set.

* `project` (string) - The name of the Openstack project you want to scope to.
  If not specified, Packer will use the environment variables
  `SDK_PROJECT` or `OS_TENANT_NAME` (in that order)

* project_id (string) - The ID of the Openstack project you want to scope to.
  If not specified, Packer will look for the `tenant_id` option, then look for the environment variables `OS_TENANT_ID` or `OS_PROJECT_ID` (in that order).

* domain (string) - The name of the Openstack domain you want to scope to.
  If not specified, Packer will use the environment variable `OS_DOMAIN_NAME`.

* domain_id (string) - The ID of the Openstack domain you want to scope to.
  If not specified, Packer will use the environment variables
  `OS_PROJECT_DOMAIN_ID` or `OS_USER_DOMAIN_ID` (in that order).

### Other Authentication Options:

* `api_key` (string) - The API key used to access OpenStack. Some OpenStack
  installations require this.
  If not specified, Packer will use the environment variables
  `SDK_API_KEY`, if set.

* `region` (string) - The name of the region, such as "DFW", in which
  to launch the server to create the AMI.
  If not specified, Packer will use the environment variables
  `SDK_REGION` or `OS_REGION_NAME` (in that order), if set.
  For a `provider` of "rackspace", it is required to specify a region,
  either using this option or with an environment variable. For other
  providers, including a private cloud, specifying a region may be optional.

* `proxy_url` (string) - For corporate networks it may be the case where 
  we want our API calls to be sent through a separate HTTP proxy than 
  external traffic.

* `insecure` (boolean) - Whether or not the connection to OpenStack can be done
  over an insecure connection. By default this is false.

### Rackspace Specific Options:

* `openstack_provider` (string) - A name of a provider that has a slightly
  different API model. Currently supported values are "openstack" (default),
  and "rackspace".

* `rackconnect_wait` (boolean) - For rackspace, whether or not to wait for
  Rackconnect to assign the machine an IP address before connecting via SSH.
  Defaults to false.

### Required:

* `flavor` (string) - The ID or full URL for the desired flavor for the
  server to be created.

* `image_name` (string) - The name of the resulting image.

* `source_image` (string) - The ID or full URL to the base image to use.
  This is the image that will be used to launch a new server and provision it.

### Optional:

* `use_floating_ip` (boolean) - Whether or not to use a floating IP for
  the instance. Defaults to false.

* `floating_ip` (string) - A specific floating IP to assign to this instance.
  `use_floating_ip` must also be set to true for this to have an affect.

* `floating_ip_pool` (string) - The name of the floating IP pool to use
  to allocate a floating IP. `use_floating_ip` must also be set to true
  for this to have an affect.

* `networks` (array of strings) - A list of networks by UUID to attach
  to this instance.

* `security_groups` (array of strings) - A list of security groups by name
  to add to this instance.

* `ssh_username` (string) - The username to use in order to communicate
  over SSH to the running server. The default is "root".

* `ssh_port` (integer) - The port that SSH will be available on. Defaults to port
  22.

* `ssh_timeout` (string) - The time to wait for SSH to become available
  before timing out. The format of this value is a duration such as "5s"
  or "1m". The default SSH timeout is "5m".

* `ssh_interface` (string) - The type of interface to connect via SSH. 
  (network pool name)
  Values useful for Rackspace are "public" or "private", and the default behavior is
  to connect via whichever is returned first from the OpenStack API.


## Basic Example: Rackspace public cloud

Here is a basic example. This is a working example to build a
Ubuntu 12.04 LTS (Precise Pangolin) on Rackspace OpenStack cloud offering.

```javascript
{
  "type": "openstack-id3",
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

## Basic Example: Private OpenStack cloud (Metacloud)

This example builds an Ubuntu 14.04 image on a private OpenStack cloud,
powered by Metacloud.

```javascript
{
  "type": "openstack-id3",
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

## Basic Example: Private / Public cloud (Generic Identity V3)

This example builds an ubuntu image on any public or private
Openstack cloud using identity version 3. Tested and working at zetta.io.
Contact your provider for details if needed.

```javascript
{
  "type": "openstack-id3",
  "provider": "https://your.provider.endpoint/v3",
  "domain": "<insert domain name>",
  "project": "<insert project name>",
  "username": "<insert username>",
  "password": "<insert password>",
  
  "image_name": "<insert name of the resulting image>",
  "source_image": "<insert image id>",
  "flavor": <insert flavor id>,
  "floating_ip_pool": "Public",
  "networks": ["<insert network id>"],
  "security_groups": ["default"],
  
  "ssh_username": "ubuntu",
}
````

Sourcing the standard OpenStack RC file obtained in the Openstack dashboard 
is recommended instead of speficying authentication options directly in the template.

## Troubleshooting


