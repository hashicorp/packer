---
description: |
    The Google Compute Image Exporter post-processor exports an image from a Packer
    googlecompute builder run and uploads it to Google Cloud Storage. The exported
    images can be easily shared and uploaded to other Google Cloud Projects.
layout: docs
page_title: 'Google Compute Image Exporter'
...

# Google Compoute Image Exporter Post-Processor

Type: `googlecompute-export`

The Google Compute Image Exporter post-processor exports the resultant image from a
googlecompute build as a gzipped tarball to Google Cloud Storage (GCS).

The exporter uses the same Google Cloud Platform (GCP) project and authentication
credentials as the googlecompute build that produced the image. A temporary VM is
started in the GCP project using these credentials. The VM mounts the built image as
a disk then dumps, compresses, and tars the image. The VM then uploads the tarball
to the provided GCS `paths` using the same credentials.

As such, the authentication credentials that built the image must have write
permissions to the GCS `paths`.


## Configuration

### Required

-   `paths` (list of string) - The list of GCS paths, e.g.
    'gs://mybucket/path/to/file.tar.gz', where the image will be exported.

### Optional

-   `keep_input_artifact` (bool) - If true, do not delete the Google Compute Engine 
    (GCE) image being exported.
    
## Basic Example

The following example builds a GCE image in the project, `my-project`, with an 
account whose keyfile is `account.json`. After the image build, a temporary VM will
be created to export the image as a gzipped tarball to 
`gs://mybucket1/path/to/file1.tar.gz` and `gs://mybucket2/path/to/file2.tar.gz`. 
`keep_input_artifact` is true, so the GCE image won't be deleted after the export.

In order for this example to work, the account associated with `account.json` must
have write access to both `gs://mybucket1/path/to/file1.tar.gz` and
`gs://mybucket2/path/to/file2.tar.gz`.

``` {.json}
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
      ]
      "keep_input_artifact": true
    }
  ]
}
```
