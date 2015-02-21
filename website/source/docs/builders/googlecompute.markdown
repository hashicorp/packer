---
layout: "docs"
page_title: "Google Compute Builder"
description: |-
  The `googlecompute` Packer builder is able to create images for use with Google Compute Engine (GCE) based on existing images. Google Compute Engine doesn't allow the creation of images from scratch.
---

# Google Compute Builder

Type: `googlecompute`

The `googlecompute` Packer builder is able to create [images](https://developers.google.com/compute/docs/images) for use with
[Google Compute Engine](https://cloud.google.com/products/compute-engine)(GCE) based on existing images. Google
Compute Engine doesn't allow the creation of images from scratch.

## Authentication

Authenticating with Google Cloud services requires at most one JSON file, 
called the _account file_. The _account file_ is **not** required if you are running
the `googlecompute` Packer builder from a GCE instance with a properly-configured
[Compute Engine Service Account](https://cloud.google.com/compute/docs/authentication).

### Running With a Compute Engine Service Account
If you run the `googlecompute` Packer builder from a GCE instance, you can configure that
instance to use a [Compute Engine Service Account](https://cloud.google.com/compute/docs/authentication). This will allow Packer to authenticate
to Google Cloud without having to bake in a separate credential/authentication file. 

To create a GCE instance that uses a service account, provide the required scopes when
launching the instance.

For `gcloud`, do this via the `--scopes` parameter:

```sh
gcloud compute --project YOUR_PROJECT instances create "INSTANCE-NAME" ... \
               --scopes "https://www.googleapis.com/auth/compute" \
                        "https://www.googleapis.com/auth/devstorage.full_control" \
               ...
```

For the [Google Developers Console](https://console.developers.google.com):

1. Choose "Show advanced options"
2. Tick "Enable Compute Engine service account"
3. Choose "Read Write" for Compute
4. Chose "Full" for "Storage"

**The service account will be used automatically by Packer as long as there is
no _account file_ specified in the Packer configuration file.**

### Running Without a Compute Engine Service Account

The [Google Developers Console](https://console.developers.google.com) allows you to
create and download a credential file that will let you use the `googlecompute` Packer
builder anywhere. To make
the process more straightforwarded, it is documented here.

1. Log into the [Google Developers Console](https://console.developers.google.com)
   and select a project.

2. Under the "APIs & Auth" section, click "Credentials."

3. Click the "Create new Client ID" button, select "Service account", and click "Create Client ID"

4. Click "Generate new JSON key" for the Service Account you just created. A JSON file will be downloaded automatically. This is your
   _account file_.

## Basic Example

Below is a fully functioning example. It doesn't do anything useful,
since no provisioners are defined, but it will effectively repackage an
existing GCE image. The account file is obtained in the previous section.

```javascript
{
  "type": "googlecompute",
  "account_file": "account.json",
  "project_id": "my-project",
  "source_image": "debian-7-wheezy-v20150127",
  "zone": "us-central1-a"
}
```

## Configuration Reference

Configuration options are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

### Required:

* `project_id` (string) - The project ID that will be used to launch instances
  and store images.

* `source_image` (string) - The source image to use to create the new image
  from. Example: `"debian-7-wheezy-v20150127"`

* `zone` (string) - The zone in which to launch the instance used to create
  the image. Example: `"us-central1-a"`

### Optional:

* `account_file` (string) - The JSON file containing your account credentials.
  Not required if you run Packer on a GCE instance with a service account.
  Instructions for creating file or using service accounts are above.

* `disk_size` (integer) - The size of the disk in GB.
  This defaults to `10`, which is 10GB.

* `image_name` (string) - The unique name of the resulting image.
  Defaults to `"packer-{{timestamp}}"`.

* `image_description` (string) - The description of the resulting image.

* `instance_name` (string) - A name to give the launched instance. Beware
  that this must be unique. Defaults to `"packer-{{uuid}}"`.

* `machine_type` (string) - The machine type. Defaults to `"n1-standard-1"`.

* `metadata` (object of key/value strings)

* `network` (string) - The Google Compute network to use for the launched
  instance. Defaults to `"default"`.

* `ssh_port` (integer) - The SSH port. Defaults to `22`.

* `ssh_timeout` (string) - The time to wait for SSH to become available.
  Defaults to `"1m"`.

* `ssh_username` (string) - The SSH username. Defaults to `"root"`.

* `state_timeout` (string) - The time to wait for instance state changes.
  Defaults to `"5m"`.

* `tags` (array of strings)

## Gotchas

Centos images have root ssh access disabled by default. Set `ssh_username` to any user, which will be created by packer with sudo access.

The machine type must have a scratch disk, which means you can't use an `f1-micro` or `g1-small` to build images.
