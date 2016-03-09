---
description: |
    The `openstack` Packer builder is able to create new images for use with
    OpenStack. The builder takes a source image, runs any provisioning necessary on
    the image after launching it, then creates a new reusable image. This reusable
    image can then be used as the foundation of new servers that are launched within
    OpenStack. The builder will create temporary keypairs that provide temporary
    access to the server while the image is being created. This simplifies
    configuration quite a bit.
layout: docs
page_title: OpenStack Builder
...

# OpenStack Builder

Type: `openstack`

The `openstack` Packer builder is able to create new images for use with
[OpenStack](http://www.openstack.org). The builder takes a source image, runs
any provisioning necessary on the image after launching it, then creates a new
reusable image. This reusable image can then be used as the foundation of new
servers that are launched within OpenStack. The builder will create temporary
keypairs that provide temporary access to the server while the image is being
created. This simplifies configuration quite a bit.

The builder does *not* manage images. Once it creates an image, it is up to you
to use it or delete it.

\~&gt; **OpenStack Liberty or later requires OpenSSL!** To use the OpenStack
builder with OpenStack Liberty (Oct 2015) or later you need to have OpenSSL
installed _if you are using temporary key pairs_, i.g. don't use
[`ssh_keypair_name`](openstack.html#ssh_keypair_name) nor
[`ssh_password`](/docs/templates/communicator.html#ssh_password). All major
OS'es have OpenSSL installed by default except Windows.

## Configuration Reference

There are many configuration options available for the builder. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `flavor` (string) - The ID, name, or full URL for the desired flavor for the
    server to be created.

-   `image_name` (string) - The name of the resulting image.

-   `identity_endpoint` (string) - The URL to the OpenStack Identity service.
    If not specified, Packer will use the environment variables `OS_AUTH_URL`,
    if set.

-   `source_image` (string) - The ID or full URL to the base image to use. This
    is the image that will be used to launch a new server and provision it.
    Unless you specify completely custom SSH settings, the source image must
    have `cloud-init` installed so that the keypair gets assigned properly.

-   `source_image_name` (string) - The name of the base image to use. This
    is an alternative way of providing `source_image` and only either of them
    can be specified.

-   `username` or `user_id` (string) - The username or id used to connect to
    the OpenStack service. If not specified, Packer will use the environment
    variable `OS_USERNAME` or `OS_USERID`, if set.

-   `password` (string) - The password used to connect to the OpenStack service.
    If not specified, Packer will use the environment variables `OS_PASSWORD`,
    if set.

### Optional:

-   `api_key` (string) - The API key used to access OpenStack. Some OpenStack
    installations require this.

-   `availability_zone` (string) - The availability zone to launch the
    server in. If this isn't specified, the default enforced by your OpenStack
    cluster will be used. This may be required for some OpenStack clusters.

-   `config_drive` (boolean) - Whether or not nova should use ConfigDrive for
     cloud-init metadata.

-   `domain_name` or `domain_id` (string) - The Domain name or ID you are
    authenticating with. OpenStack installations require this if identity v3 is used.
    Packer will use the environment variable `OS_DOMAIN_NAME` or `OS_DOMAIN_ID`, if set.

-   `endpoint_type` (string) - The endpoint type to use. Can be any of "internal",
    "internalURL", "admin", "adminURL", "public", and "publicURL". By default
    this is "public".

-   `floating_ip` (string) - A specific floating IP to assign to this instance.

-   `floating_ip_pool` (string) - The name of the floating IP pool to use to
    allocate a floating IP.

-   `insecure` (boolean) - Whether or not the connection to OpenStack can be
    done over an insecure connection. By default this is false.

-   `metadata` (object of key/value strings) - Glance metadata that will be
    applied to the image.

-   `networks` (array of strings) - A list of networks by UUID to attach to
    this instance.

-   `rackconnect_wait` (boolean) - For rackspace, whether or not to wait for
    Rackconnect to assign the machine an IP address before connecting via SSH.
    Defaults to false.

-   `region` (string) - The name of the region, such as "DFW", in which to
    launch the server to create the AMI. If not specified, Packer will use the
    environment variable `OS_REGION_NAME`, if set.

-   `security_groups` (array of strings) - A list of security groups by name to
    add to this instance.

-   `ssh_interface` (string) - The type of interface to connect via SSH. Values
    useful for Rackspace are "public" or "private", and the default behavior is
    to connect via whichever is returned first from the OpenStack API.

-   `ssh_ip_version` (string) - The IP version to use for SSH connections, valid
    values are `4` and `6`. Useful on dual stacked instances where the default
    behaviour is to connect via whichever IP address is returned first from the
    OpenStack API.

-   `ssh_keypair_name` (string) - If specified, this is the key that will be
    used for SSH with the machine. By default, this is blank, and Packer will
    generate a temporary keypair.
    [`ssh_private_key_file`](/docs/templates/communicator.html#ssh_private_key_file)
    must be specified with this.

-   `tenant_id` or `tenant_name` (string) - The tenant ID or name to boot the
    instance into. Some OpenStack installations require this. If not specified,
    Packer will use the environment variable `OS_TENANT_NAME`, if set. Tenant
    is also called Project in later versions of OpenStack.

-   `use_floating_ip` (boolean) - _Deprecated_ use `floating_ip` or `floating_ip_pool`
    instead.

-   `user_data` (string) - User data to apply when launching the instance. Note
    that you need to be careful about escaping characters due to the templates
    being JSON. It is often more convenient to use `user_data_file`, instead.

-   `user_data_file` (string) - Path to a file that will be used for the user
    data when launching the instance.

## Basic Example: DevStack

Here is a basic example. This is a example to build on DevStack running in a VM.

``` {.javascript}
{
  "type": "openstack",
  "identity_endpoint": "http://<destack-ip>:5000/v3",
  "tenant_name": "admin",
  "domain_name": "Default",
  "username": "admin",
  "password": "<your admin password>",
  "region": "RegionOne",
  "ssh_username": "root",
  "image_name": "Test image",
  "source_image": "<image id>",
  "flavor": "m1.tiny",
  "insecure": "true"
}

```

## Basic Example: Rackspace public cloud

Here is a basic example. This is a working example to build a Ubuntu 12.04 LTS
(Precise Pangolin) on Rackspace OpenStack cloud offering.

``` {.javascript}
{
  "type": "openstack",
  "username": "foo",
  "password": "foo",
  "region": "DFW",
  "ssh_username": "root",
  "image_name": "Test image",
  "source_image": "23b564c9-c3e6-49f9-bc68-86c7a9ab5018",
  "flavor": "2"
}
```

## Basic Example: Private OpenStack cloud

This example builds an Ubuntu 14.04 image on a private OpenStack cloud, powered
by Metacloud.

``` {.javascript}
{
  "type": "openstack",
  "ssh_username": "root",
  "image_name": "ubuntu1404_packer_test_1",
  "source_image": "91d9c168-d1e5-49ca-a775-3bfdbb6c97f1",
  "flavor": "2"
}
```

In this case, the connection information for connecting to OpenStack doesn't
appear in the template. That is because I source a standard OpenStack script
with environment variables set before I run this. This script is setting
environment variables like:

-   `OS_AUTH_URL`
-   `OS_TENANT_ID`
-   `OS_USERNAME`
-   `OS_PASSWORD`

This is slightly different when identity v3 is used:

-   `OS_AUTH_URL`
-   `OS_USERNAME`
-   `OS_PASSWORD`
-   `OS_DOMAIN_NAME`
-   `OS_TENANT_NAME`

This will authenticate the user on the domain and scope you to the project.
A tenant is the same as a project. It's optional to use names or IDs in v3.
This means you can use `OS_USERNAME` or `OS_USERID`,  `OS_TENANT_ID` or
`OS_TENANT_NAME` and `OS_DOMAIN_ID` or `OS_DOMAIN_NAME`.

The above example would be equivalent to an RC file looking like this :

    export OS_AUTH_URL="https://identity.myprovider/v3"
    export OS_USERNAME="myuser"
    export OS_PASSWORD="password"
    export OS_USER_DOMAIN_NAME="mydomain"
    export OS_PROJECT_DOMAIN_NAME="mydomain"

## Notes on OpenStack Authorization

The simplest way to get all settings for authorization agains OpenStack is to
go into the OpenStack Dashboard (Horizon) select your _Project_ and navigate
_Project, Access & Security_, select _API Access_ and _Download OpenStack RC
File v3_. Source the file, and select your wanted region by setting
environment variable `OS_REGION_NAME` or `OS_REGION_ID` and `export
OS_TENANT_NAME=$OS_PROJECT_NAME` or `export OS_TENANT_ID=$OS_PROJECT_ID`.

\~&gt; `OS_TENANT_NAME` or `OS_TENANT_ID` must be used even with Identity v3,
`OS_PROJECT_NAME` and `OS_PROJECT_ID` has no effect in Packer.

To troubleshoot authorization issues test you environment variables with the
OpenStack cli. It can be installed with
```pip install --user python-openstackclient```.
