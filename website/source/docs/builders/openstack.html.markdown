---
layout: "docs"
---

# OpenStack Builder

Type: `openstack`

The `openstack` builder is able to create new images for use with
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

Required:

* `flavor` (string) - The ID or full URL for the desired flavor for the
  server to be created.

* `image_name` (string) - The name of the resulting image.

* `password` (string) - The password used to connect to the OpenStack service.
  If not specified, Packer will attempt to read this from the
  `SDK_PASSWORD` or `OS_PASSWORD` environment variable.

* `provider` (string) - The provider used to connect to the OpenStack service.
  If not specified, Packer will attempt to read this from the
  `SDK_PROVIDER` environment variable. For Rackspace this should be `rackspace-us`
  or `rackspace-uk`.

* `region` (string) - The name of the region, such as "DFW", in which
  to launch the server to create the AMI. If not specified, Packer will
  attempt to read this from the `SDK_REGION` or `OS_REGION_NAME` environmental
  variables.

* `source_image` (string) - The ID or full URL to the base image to use.
  This is the image that will be used to launch a new server and provision it.

* `username` (string) - The username used to connect to the OpenStack service.
  If not specified, Packer will attempt to read this from the
  `SDK_USERNAME` or `OS_USERNAME` environment variable.

Optional:

* `api_key` (string) - The API key used to access OpenStack. Some OpenStack
  installations require this. If not specified, Packer will attempt to
  read this from the `SDK_API_KEY` environmental variable.

* `project` (string) - The project name to boot the instance into. Some
  OpenStack installations require this. If not specified, Packer will attempt
  to read this from the `SDK_PROJECT` or `OS_TENANT_NAME` environmental
  variables.

* `provider` (string) - A name of a provider that has a slightly
  different API model. Currently supported values are "openstack" (default),
  and "rackspace". If not specified, Packer will attempt to read this from
  the `SDK_PROVIDER` or `OS_AUTH_URL` environmental variables.

* `ssh_port` (int) - The port that SSH will be available on. Defaults to port
  22.

* `ssh_timeout` (string) - The time to wait for SSH to become available
  before timing out. The format of this value is a duration such as "5s"
  or "1m". The default SSH timeout is "5m".

* `ssh_username` (string) - The username to use in order to communicate
  over SSH to the running server. The default is "root".

## Basic Example

Here is a basic example. This is a working example to build a
Ubuntu 12.04 LTS (Precise Pangolin) on Rackspace OpenStack cloud offering.

<pre class="prettyprint">
{
  "type": "openstack",
  "username": "",
  "password": "",
  "provider": "rackspace-us",
  "region": "DFW",
  "ssh_username": "root",
  "image_name": "Test image",
  "source_image": "23b564c9-c3e6-49f9-bc68-86c7a9ab5018",
  "flavor": "2"
}
</pre>

## Troubleshooting

*I get the error "Missing or incorrect provider"*

* Verify your "username", "password" and "provider" settings.

*I get the error "Missing endpoint, or insufficient privileges to access endpoint"*

* Verify your "region" setting.
