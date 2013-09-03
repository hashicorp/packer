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
  `SDK_PASSWORD` environment variable.

* `provider` (string) - The provider used to connect to the OpenStack service.
  If not specified, Packer will attempt to read this from the
  `SDK_PROVIDER` environment variable.

* `region` (string) - The name of the region, such as "DFW", in which
  to launch the server to create the AMI.

* `source_image` (string) - The ID or full URL to the base image to use.
  This is the image that will be used to launch a new server and provision it.

* `username` (string) - The username used to connect to the OpenStack service.
  If not specified, Packer will attempt to read this from the
  `SDK_USERNAME` environment variable.

Optional:

* `project` (string) - The project name to boot the instance into. Some
  OpenStack installations require this. By default this is empty.

* `ssh_port` (int) - The port that SSH will be available on. Defaults to port
  22.

* `ssh_timeout` (string) - The time to wait for SSH to become available
  before timing out. The format of this value is a duration such as "5s"
  or "5m". The default SSH timeout is "1m".

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
  "provider": "",
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
