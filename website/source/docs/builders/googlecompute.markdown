---
layout: "docs"
---

# Google Compute Builder

Type: `googlecompute`

The `googlecompute` builder is able to create new [images](https://developers.google.com/compute/docs/images)
for use with [Google Compute Engine](https://cloud.google.com/products/compute-engine).

## Basic Example

Below is a fully functioning example. It doesn't do anything useful, since no provisioners are defined, but it will effectively repackage an GCE image.

<pre class="prettyprint">
{
  "type": "googlecompute",
  "bucket_name": "packer-images",
  "client_secrets_file": "client_secret_XXXXXX-XXXXXX.apps.googleusercontent.com.json",
  "private_key_file": "XXXXXX-privatekey.pem",
  "project_id": "my-project",
  "source_image": "debian-7-wheezy-v20131014",
  "zone": "us-central1-a"
}
</pre>

## Configuration Reference

Configuration options are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

Required:

* `bucket_name` (string) - The Google Cloud Storage bucket to store images.
* `client_secrets_file` (string) - The client secrets file.
* `private_key_file` (string) - The service account private key.
* `project_id` (string) - The GCE project id.
* `source_image` (string) - The source image. Example `debian-7-wheezy-v20131014`.
* `zone` (string) - The GCE zone.

Optional:

* `image_name` (string) - The unique name of the resulting image. Defaults to `packer-{{timestamp}}`.
* `image_description` (string) - The description of the resulting image.
* `machine_type` (string) - The machine type. Defaults to `n1-standard-1`.
* `network` (string) - The Google Compute network. Defaults to `default`.
* `passphrase` (string) - The passphrase to use if the `private_key_file` is encrypted.
* `ssh_port` (int) - The SSH port. Defaults to `22`.
* `ssh_timeout` (string) - The time to wait for SSH to become available. Defaults to `1m`.
* `ssh_username` (string) - The SSH username. Defaults to `root`.
* `state_timeout` (string) - The time to wait for instance state changes. Defaults to `5m`.

## Gotchas

Centos images have root ssh access disabled by default. Set `ssh_username` to any user, which will be created by packer with sudo access.

The machine type must have a scratch disk, which means you can't use an `f1-micro` or `g1-small` to build images.
