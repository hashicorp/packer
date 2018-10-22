---
description: |
    The googlecompute Packer builder is able to create images for use with
    Google Cloud Compute Engine (GCE) based on existing images.
layout: docs
page_title: 'Google Compute - Builders'
sidebar_current: 'docs-builders-googlecompute'
---

# Google Compute Builder

Type: `googlecompute`

The `googlecompute` Packer builder is able to create
[images](https://developers.google.com/compute/docs/images) for use with [Google
Compute Engine](https://cloud.google.com/products/compute-engine) (GCE) based on
existing images.

It is possible to build images from scratch, but not with the `googlecompute` Packer builder.
The process is recommended only for advanced users, please see [Building GCE Images from Scratch]
(https://cloud.google.com/compute/docs/tutorials/building-images)
and the [Google Compute Import Post-Processor](/docs/post-processors/googlecompute-import.html)
for more information.

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

``` shell
$ gcloud compute --project YOUR_PROJECT instances create "INSTANCE-NAME" ... \
    --scopes "https://www.googleapis.com/auth/compute,https://www.googleapis.com/auth/devstorage.full_control" \
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

2.  Under the "API Manager" section, click "Credentials."

3.  Click the "Create credentials" button, select "Service account key"

4.  Create a new service account that at least has `Compute Engine Instance Admin (v1)` and `Service Account User` roles.

5.  Choose `JSON` as the Key type and click "Create".
    A JSON file will be downloaded automatically. This is your *account file*.

### Precedence of Authentication Methods

Packer looks for credentials in the following places, preferring the first
location found:

1.  An `account_file` option in your packer file.

2.  A JSON file (Service Account) whose path is specified by the
`GOOGLE_APPLICATION_CREDENTIALS` environment variable.

3.  A JSON file in a location known to the `gcloud` command-line tool.
(`gcloud` creates it when it's configured)

    On Windows, this is:

        %APPDATA%/gcloud/application_default_credentials.json

    On other systems:

        $HOME/.config/gcloud/application_default_credentials.json

4.  On Google Compute Engine and Google App Engine Managed VMs, it fetches
credentials from the metadata server. (Needs a correct VM authentication scope
configuration, see above.)

## Examples

### Basic Example

Below is a fully functioning example. It doesn't do anything useful since no
provisioners or startup-script metadata are defined, but it will effectively
repackage an existing GCE image. The account\_file is obtained in the previous
section. If it parses as JSON it is assumed to be the file itself, otherwise, it
is assumed to be the path to the file containing the JSON.

``` json
{
  "builders": [
    {
      "type": "googlecompute",
      "account_file": "account.json",
      "project_id": "my project",
      "source_image": "debian-7-wheezy-v20150127",
      "ssh_username": "packer",
      "zone": "us-central1-a"
    }
  ]
}
```

### Windows Example

Before you can provision using the winrm communicator, you need to allow traffic
through google's firewall on the winrm port (tcp:5986).
You can do so using the gcloud command.
```
gcloud compute firewall-rules create allow-winrm --allow tcp:5986
```
Or alternatively by navigating to https://console.cloud.google.com/networking/firewalls/list.

Once this is set up, the following is a complete working packer config after
setting a valid `account_file` and `project_id`:

``` {.json}
{
  "builders": [
    {
      "type": "googlecompute",
      "account_file": "account.json",
      "project_id": "my project",
      "source_image": "windows-server-2016-dc-v20170227",
      "disk_size": "50",
      "machine_type": "n1-standard-1",
      "communicator": "winrm",
      "winrm_username": "packer_user",
      "winrm_insecure": true,
      "winrm_use_ssl": true,
      "metadata": {
        "windows-startup-script-cmd": "winrm quickconfig -quiet & net user /add packer_user & net localgroup administrators packer_user /add & winrm set winrm/config/service/auth @{Basic=\"true\"}"
      },
      "zone": "us-central1-a"
    }
  ]
}
```
This build can take up to 15 min.

### Nested Hypervisor Example

This is an example of using the `image_licenses` configuration option to create a GCE image that has nested virtualization enabled. See
[Enabling Nested Virtualization for VM Instances](https://cloud.google.com/compute/docs/instances/enable-nested-virtualization-vm-instances)
for details.

``` json
{
  "builders": [
    {
      "type": "googlecompute",
      "account_file": "account.json",
      "project_id": "my project",
      "source_image_family": "centos-7",
      "ssh_username": "packer",
      "zone": "us-central1-a",
      "image_licenses": ["projects/vm-options/global/licenses/enable-vmx"]
    }
  ]
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

-   `source_image` (string) - The source image to use to create the new image
    from. You can also specify `source_image_family` instead. If both
    `source_image` and `source_image_family` are specified, `source_image`
    takes precedence. Example: `"debian-8-jessie-v20161027"`

-   `source_image_family` (string) - The source image family to use to create
    the new image from. The image family always returns its latest image that
    is not deprecated. Example: `"debian-8"`.

-   `zone` (string) - The zone in which to launch the instance used to create
    the image. Example: `"us-central1-a"`

### Optional:

-   `account_file` (string) - The JSON file containing your account credentials.
    Not required if you run Packer on a GCE instance with a service account.
    Instructions for creating the file or using service accounts are above.

-   `accelerator_count` (number) - Number of guest accelerator cards to add to the launched instance.

-   `accelerator_type` (string) - Full or partial URL of the guest accelerator type. GPU accelerators can only be used with
    `"on_host_maintenance": "TERMINATE"` option set.
    Example: `"projects/project_id/zones/europe-west1-b/acceleratorTypes/nvidia-tesla-k80"`

-   `address` (string) - The name of a pre-allocated static external IP address.
    Note, must be the name and not the actual IP address.

-   `disable_default_service_account` (bool) - If true, the default service account will not be used if `service_account_email`
    is not specified. Set this value to true and omit `service_account_email` to provision a VM with no service account.

-   `disk_name` (string) - The name of the disk, if unset the instance name will be
    used.

-   `disk_size` (number) - The size of the disk in GB. This defaults to `10`,
    which is 10GB.

-   `disk_type` (string) - Type of disk used to back your instance, like `pd-ssd` or `pd-standard`. Defaults to `pd-standard`.

-   `image_description` (string) - The description of the resulting image.

-   `image_family` (string) - The name of the image family to which the
    resulting image belongs. You can create disks by specifying an image family
    instead of a specific image name. The image family always returns its
    latest image that is not deprecated.

-   `image_labels` (object of key/value strings) - Key/value pair labels to
    apply to the created image.

-   `image_licenses` (array of strings) - Licenses to apply to the created image.

-   `image_name` (string) - The unique name of the resulting image. Defaults to
    `"packer-{{timestamp}}"`.

-   `instance_name` (string) - A name to give the launched instance. Beware that
    this must be unique. Defaults to `"packer-{{uuid}}"`.

-   `labels` (object of key/value strings) - Key/value pair labels to apply to
    the launched instance.

-   `machine_type` (string) - The machine type. Defaults to `"n1-standard-1"`.

-   `metadata` (object of key/value strings) - Metadata applied to the launched
    instance.

-   `min_cpu_platform` (string) - A Minimum CPU Platform for VM Instance.
    Availability and default CPU platforms vary across zones, based on 
    the hardware available in each GCP zone. [Details](https://cloud.google.com/compute/docs/instances/specify-min-cpu-platform)

-   `network` (string) - The Google Compute network id or URL to use for the
    launched instance. Defaults to `"default"`. If the value is not a URL, it
    will be interpolated to `projects/((network_project_id))/global/networks/((network))`.
    This value is not required if a `subnet` is specified.


-   `network_project_id` (string) - The project ID for the network and subnetwork
    to use for launched instance. Defaults to `project_id`.

-   `omit_external_ip` (boolean) - If true, the instance will not have an external IP.
    `use_internal_ip` must be true if this property is true.

-   `on_host_maintenance` (string) - Sets Host Maintenance Option. Valid
    choices are `MIGRATE` and `TERMINATE`. Please see [GCE Instance Scheduling
    Options](https://cloud.google.com/compute/docs/instances/setting-instance-scheduling-options),
    as not all machine\_types support `MIGRATE` (i.e. machines with GPUs).
    If preemptible is true this can only be `TERMINATE`. If preemptible
    is false, it defaults to `MIGRATE`

-   `preemptible` (boolean) - If true, launch a preemptible instance.

-   `region` (string) - The region in which to launch the instance. Defaults to
    the region hosting the specified `zone`.

-   `service_account_email` (string) - The service account to be used for launched instance. Defaults to
    the project's default service account unless `disable_default_service_account` is true.

-   `scopes` (array of strings) - The service account scopes for launched instance.
    Defaults to:

    ``` json
    [
      "https://www.googleapis.com/auth/userinfo.email",
      "https://www.googleapis.com/auth/compute",
      "https://www.googleapis.com/auth/devstorage.full_control"
    ]
    ```

-   `source_image_project_id` (string) - The project ID of the
    project containing the source image.

-   `startup_script_file` (string) - The path to a startup script to run on
    the VM from which the image will be made.

-   `state_timeout` (string) - The time to wait for instance state changes.
    Defaults to `"5m"`.

-   `subnetwork` (string) - The Google Compute subnetwork id or URL to use for
    the launched instance. Only required if the `network` has been created with
    custom subnetting. Note, the region of the subnetwork must match the `region`
    or `zone` in which the VM is launched. If the value is not a URL, it
    will be interpolated to `projects/((network_project_id))/regions/((region))/subnetworks/((subnetwork))`


-   `tags` (array of strings) - Assign network tags to apply firewall rules to
    VM instance.

-   `use_internal_ip` (boolean) - If true, use the instance's internal IP
    instead of its external IP during building.

## Startup Scripts

Startup scripts can be a powerful tool for configuring the instance from which the image is made.
The builder will wait for a startup script to terminate. A startup script can be provided via the
`startup_script_file` or `startup-script` instance creation `metadata` field. Therefore, the build
time will vary depending on the duration of the startup script. If `startup_script_file` is set,
the `startup-script` `metadata` field will be overwritten. In other words, `startup_script_file`
takes precedence.

The builder does not check for a pass/fail/error signal from the startup script, at this time. Until
such support is implemented, startup scripts should be robust, as an image will still be built even
when a startup script fails.

### Windows

A Windows startup script can only be provided via the `windows-startup-script-cmd` instance
creation `metadata` field. The builder will *not* wait for a Windows startup script to
terminate. You have to ensure that it finishes before the instance shuts down.

### Logging

Startup script logs can be copied to a Google Cloud Storage (GCS) location specified via the
`startup-script-log-dest` instance creation `metadata` field. The GCS location must be writeable by
the credentials provided in the builder config's `account_file`.

## Gotchas

CentOS and recent Debian images have root ssh access disabled by default. Set `ssh_username` to
any user, which will be created by packer with sudo access.

The machine type must have a scratch disk, which means you can't use an
`f1-micro` or `g1-small` to build images.
