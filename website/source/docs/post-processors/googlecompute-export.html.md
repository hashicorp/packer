---
description: |
    The Google Compute Image Exporter post-processor exports an image from a Packer
    googlecompute builder run and uploads it to Google Cloud Storage. The exported
    images can be easily shared and uploaded to other Google Cloud Projects.
layout: docs
page_title: 'Google Compute Image Exporter - Post-Processors'
sidebar_current: 'docs-post-processors-googlecompute-export'
---

# Google Compute Image Exporter Post-Processor

Type: `googlecompute-export`

The Google Compute Image Exporter post-processor exports the resultant image
from a googlecompute build as a gzipped tarball to Google Cloud Storage (GCS).

The exporter uses the same Google Cloud Platform (GCP) project and
authentication credentials as the googlecompute build that produced the image.
A temporary VM is started in the GCP project using these credentials. The VM
mounts the built image as a disk then dumps, compresses, and tars the image.
The VM then uploads the tarball to the provided GCS `paths` using the same
credentials.

As such, the authentication credentials that built the image must have write
permissions to the GCS `paths`.

## Configuration

### Required

-   `paths` (list of string) - The list of GCS paths, e.g.
    'gs://mybucket/path/to/file.tar.gz', where the image will be exported.

### Optional

-   `account_file` (string) - The JSON file containing your account
    credentials. If specified, this take precedence over `googlecompute`
    builder authentication method.

-   `disk_size` (number) - The size of the export instances disk, this disk
    is unused for the export but a larger size increase `pd-ssd` read speed.
    This defaults to `200`, which is 200GB.

-   `disk_type` (string) - Type of disk used to back export instance, like
    `pd-ssd` or `pd-standard`. Defaults to `pd-ssd`.

-   `keep_input_artifact` (boolean) - If true, do not delete the Google Compute
    Engine (GCE) image being exported.

-   `machine_type` (string) - The export instance machine type. Defaults
    to `"n1-highcpu-4"`.

-   `network` (string) - The Google Compute network id or URL to use for the
    export instance. Defaults to `"default"`. If the value is not a URL, it
    will be interpolated to
    `projects/((network_project_id))/global/networks/((network))`. This value
    is not required if a `subnet` is specified.

-   `subnetwork` (string) - The Google Compute subnetwork id or URL to use for
    the export instance. Only required if the `network` has been created with
    custom subnetting. Note, the region of the subnetwork must match the
    `zone` in which the VM is launched. If the value is not a URL,
    it will be interpolated to
    `projects/((network_project_id))/regions/((region))/subnetworks/((subnetwork))`

-   `zone` (string) - The zone in which to launch the export instance. Defaults
    to `googlecompute` builder zone. Example: `"us-central1-a"`

## Basic Example

The following example builds a GCE image in the project, `my-project`, with an
account whose keyfile is `account.json`. After the image build, a temporary VM
will be created to export the image as a gzipped tarball to
`gs://mybucket1/path/to/file1.tar.gz` and
`gs://mybucket2/path/to/file2.tar.gz`. `keep_input_artifact` is true, so the
GCE image won't be deleted after the export.

In order for this example to work, the account associated with `account.json`
must have write access to both `gs://mybucket1/path/to/file1.tar.gz` and
`gs://mybucket2/path/to/file2.tar.gz`.

``` json
{
  "builders": [
    {
      "type": "googlecompute",
      "account_file": "account.json",
      "project_id": "my-project",
      "source_image": "debian-7-wheezy-v20150127",
      "zone": "us-central1-a"
    }
  ],
  "post-processors": [
    {
      "type": "googlecompute-export",
      "paths": [
        "gs://mybucket1/path/to/file1.tar.gz",
        "gs://mybucket2/path/to/file2.tar.gz"
      ],
      "keep_input_artifact": true
    }
  ]
}
```
