---
description: |
    The `googlecompute` Packer builder is able to create images for use with Google
    Compute Engine (GCE) based on existing images. Building GCE images from scratch
    is not possible from Packer at this time. For building images from scratch, please see
    [Building GCE Images from Scratch](https://cloud.google.com/compute/docs/tutorials/building-images).
layout: docs
page_title: Google Compute Builder
...

# Google Compute Builder

Type: `googlecompute`

The `googlecompute` Packer builder is able to create
[images](https://developers.google.com/compute/docs/images) for use with [Google
Compute Engine](https://cloud.google.com/products/compute-engine)(GCE) based on
existing images. Building GCE images from scratch is not possible from Packer at
this time. For building images from scratch, please see
[Building GCE Images from Scratch](https://cloud.google.com/compute/docs/tutorials/building-images).
## Authentication

Authenticating with Google Cloud services requires at most one JSON file, called
the *account file*. The *account file* is **not** required if you are running
the `googlecompute` Packer builder from a GCE instance with a
properly-configured [Compute Engine Service
Account](https://cloud.google.com/compute/docs/authentication).

### Running With a Compute Engine Service Account

If you run the `googlecompute` Packer builder from a GCE instance, you can
configure that instance to use a [Compute Engine Service
Account](https://cloud.google.com/compute/docs/authentication). This will allow
Packer to authenticate to Google Cloud without having to bake in a separate
credential/authentication file.

To create a GCE instance that uses a service account, provide the required
scopes when launching the instance.

For `gcloud`, do this via the `--scopes` parameter:

``` {.sh}
gcloud compute --project YOUR_PROJECT instances create "INSTANCE-NAME" ... \
               --scopes "https://www.googleapis.com/auth/compute" \
                        "https://www.googleapis.com/auth/devstorage.full_control" \
               ...
```

For the [Google Developers Console](https://console.developers.google.com):

1.  Choose "Show advanced options"
2.  Tick "Enable Compute Engine service account"
3.  Choose "Read Write" for Compute
4.  Chose "Full" for "Storage"

**The service account will be used automatically by Packer as long as there is
no *account file* specified in the Packer configuration file.**

### Running Without a Compute Engine Service Account

The [Google Developers Console](https://console.developers.google.com) allows
you to create and download a credential file that will let you use the
`googlecompute` Packer builder anywhere. To make the process more
straightforwarded, it is documented here.

1.  Log into the [Google Developers
    Console](https://console.developers.google.com) and select a project.

2.  Under the "APIs & Auth" section, click "Credentials."

3.  Click the "Create new Client ID" button, select "Service account", and click
    "Create Client ID"

4.  Click "Generate new JSON key" for the Service Account you just created. A
    JSON file will be downloaded automatically. This is your *account file*.

## Basic Example

Below is a fully functioning example. It doesn't do anything useful, since no
provisioners or startup-script metadata are defined, but it will effectively
repackage an existing GCE image. The account_file is obtained in the previous
section. If it parses as JSON it is assumed to be the file itself, otherwise it
is assumed to be the path to the file containing the JSON.

``` {.javascript}
{
  "builders": [{
    "type": "googlecompute",
    "account_file": "account.json",
    "project_id": "my project",
    "source_image": "debian-7-wheezy-v20150127",
    "zone": "us-central1-a"
  }]
}
```

## Configuration Reference

Configuration options are organized below into two categories: required and
optional. Within each category, the available options are alphabetized and
described.

In addition to the options listed here, a
[communicator](/docs/templates/communicator.html) can be configured for this
builder.

### Required:

-   `project_id` (string) - The project ID that will be used to launch instances
    and store images.

-   `source_image` (string) - The source image to use to create the new
    image from. Example: `"debian-7-wheezy-v20150127"`

-   `zone` (string) - The zone in which to launch the instance used to create
    the image. Example: `"us-central1-a"`

### Optional:

-   `account_file` (string) - The JSON file containing your account credentials.
    Not required if you run Packer on a GCE instance with a service account.
    Instructions for creating file or using service accounts are above.

-   `address` (string) - The name of a pre-allocated static external IP address.
    Note, must be the name and not the actual IP address.

-   `disk_size` (integer) - The size of the disk in GB. This defaults to `10`,
    which is 10GB.

-   `disk_type` (string) - Type of disk used to back your instance, like `pd-ssd` or `pd-standard`. Defaults to `pd-standard`.

-   `image_description` (string) - The description of the resulting image.

-   `image_family` (string) - The name of the image family to which the resulting image belongs. You can create disks by specifying an image family instead of a specific image name. The image family always returns its latest image that is not deprecated.

-   `image_name` (string) - The unique name of the resulting image. Defaults to
    `"packer-{{timestamp}}"`.

-   `instance_name` (string) - A name to give the launched instance. Beware that
    this must be unique. Defaults to `"packer-{{uuid}}"`.

-   `machine_type` (string) - The machine type. Defaults to `"n1-standard-1"`.

-   `metadata` (object of key/value strings)

-   `network` (string) - The Google Compute network to use for the
    launched instance. Defaults to `"default"`.

-   `omit_external_ip` (boolean) - If true, the instance will not have an external IP.
    `use_internal_ip` must be true if this property is true.

-   `preemptible` (boolean) - If true, launch a preembtible instance.

-   `region` (string) - The region in which to launch the instance. Defaults to
    to the region hosting the specified `zone`.

-   `startup_script_file` (string) - The filepath to a startup script to run on 
    the VM from which the image will be made.

-   `state_timeout` (string) - The time to wait for instance state changes.
    Defaults to `"5m"`.

-   `subnetwork` (string) - The Google Compute subnetwork to use for the launced
     instance. Only required if the `network` has been created with custom
     subnetting.
     Note, the region of the subnetwork must match the `region` or `zone` in
     which the VM is launched.

-   `tags` (array of strings)

-   `use_internal_ip` (boolean) - If true, use the instance's internal IP
    instead of its external IP during building.
    
## Startup Scripts

Startup scripts can be a powerful tool for configuring the instance from which the image is made. 
The builder will wait for a startup script to terminate. A startup script can be provided via the
`startup_script_file` or 'startup-script' instance creation `metadata` field. Therefore, the build
time will vary depending on the duration of the startup script. If `startup_script_file` is set,
the 'startup-script' `metadata` field will be overwritten. In other words,`startup_script_file`
takes precedence.

The builder does not check for a pass/fail/error signal from the startup script, at this time. Until
such support is implemented, startup scripts should be robust, as an image will still be built even
when a startup script fails.

### Windows
Startup scripts do not work on Windows builds, at this time.

### Logging
Startup script logs can be copied to a Google Cloud Storage (GCS) location specified via the
'startup-script-log-dest' instance creation `metadata` field. The GCS location must be writeable by
the credentials provided in the builder config's `account_file`.

## Gotchas

Centos and recent Debian images have root ssh access disabled by default. Set `ssh_username` to
any user, which will be created by packer with sudo access.

The machine type must have a scratch disk, which means you can't use an
`f1-micro` or `g1-small` to build images.
