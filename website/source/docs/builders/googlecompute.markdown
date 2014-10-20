---
layout: "docs"
page_title: "Google Compute Builder"
---

# Google Compute Builder

Type: `googlecompute`

The `googlecompute` builder is able to create
[images](https://developers.google.com/compute/docs/images)
for use with [Google Compute Engine](https://cloud.google.com/products/compute-engine)
(GCE) based on existing images. Google Compute Engine doesn't allow the creation
of images from scratch.

## Authentication

Authenticating with Google Cloud services requires two separate JSON
files: one which we call the _account file_ and the _client secrets file_.

Both of these files are downloaded directly from the
[Google Developers Console](https://console.developers.google.com). To make
the process more straightforwarded, it is documented here.

1. Log into the [Google Developers Console](https://console.developers.google.com)
   and select a project.

2. Under the "APIs & Auth" section, click "Credentials."

3. Click the "Download JSON" button under the "Compute Engine and App Engine"
   account in the OAuth section. The file should start with "client\_secrets".
   This is your _client secrets file_.

4. Create a new OAuth client ID and select "Service Account" as the type
   of account. Once created, a JSON file should be downloaded. This is your
   _account file_.

## Basic Example

Below is a fully functioning example. It doesn't do anything useful,
since no provisioners are defined, but it will effectively repackage an
existing GCE image. The client secrets file and private key file are the
files obtained in the previous section.

```javascript
{
  "type": "googlecompute",
  "bucket_name": "my-project-packer-images",
  "account_file": "account.json",
  "client_secrets_file": "client_secret.json",
  "project_id": "my-project",
  "source_image": "debian-7-wheezy-v20140718",
  "zone": "us-central1-a"
}
```

## Configuration Reference

Configuration options are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

### Required:

* `account_file` (string) - The JSON file containing your account credentials.
     Instructions for how to retrieve these are above.

* `bucket_name` (string) - The Google Cloud Storage bucket to store the
  images that are created. The bucket must already exist in your project.

* `client_secrets_file` (string) - The client secrets JSON file that
  was set up in the section above.

* `private_key_file` (string) - The client private key file that was
  generated in the section above.

* `project_id` (string) - The project ID that will be used to launch instances
  and store images.

* `source_image` (string) - The source image to use to create the new image
  from. Example: "debian-7"

* `zone` (string) - The zone in which to launch the instance used to create
  the image. Example: "us-central1-a"

### Optional:

* `disk_size` (integer) - The size of the disk in GB.
  This defaults to 10, which is 10GB.

* `image_name` (string) - The unique name of the resulting image.
  Defaults to `packer-{{timestamp}}`.

* `image_description` (string) - The description of the resulting image.

* `instance_name` (string) - A name to give the launched instance. Beware
  that this must be unique. Defaults to "packer-{{uuid}}".

* `machine_type` (string) - The machine type. Defaults to `n1-standard-1`.

* `metadata` (object of key/value strings)

* `network` (string) - The Google Compute network to use for the launched
  instance. Defaults to `default`.

* `ssh_port` (integer) - The SSH port. Defaults to 22.

* `ssh_timeout` (string) - The time to wait for SSH to become available.
  Defaults to "1m".

* `ssh_username` (string) - The SSH username. Defaults to "root".

* `state_timeout` (string) - The time to wait for instance state changes.
  Defaults to "5m".

* `tags` (array of strings)

## Gotchas

Centos images have root ssh access disabled by default. Set `ssh_username` to any user, which will be created by packer with sudo access.

The machine type must have a scratch disk, which means you can't use an `f1-micro` or `g1-small` to build images.
