---
description: |
    The Google Compute Image Import post-processor takes a compressed raw disk
    image and imports it to a GCE image available to Google Compute Engine.
layout: docs
page_title: 'Google Compute Image Import - Post-Processors'
sidebar_current: 'docs-post-processors-googlecompute-import'
---

# Google Compute Image Import Post-Processor

Type: `googlecompute-import`

The Google Compute Image Import post-processor takes a compressed raw disk
image and imports it to a GCE image available to Google Compute Engine.

~&gt; This post-processor is for advanced users. Please ensure you read the
[GCE import
documentation](https://cloud.google.com/compute/docs/images/import-existing-image)
before using this post-processor.

## How Does it Work?

The import process operates by uploading a temporary copy of the compressed raw
disk image to a GCS bucket, and calling an import task in GCP on the raw disk
file. Once completed, a GCE image is created containing the converted virtual
machine. The temporary raw disk image copy in GCS can be discarded after the
import is complete.

Google Cloud has very specific requirements for images being imported. Please
see the [GCE import
documentation](https://cloud.google.com/compute/docs/images/import-existing-image)
for details.

## Configuration

### Required

-   `account_file` (string) - The JSON file containing your account
    credentials.

-   `bucket` (string) - The name of the GCS bucket where the raw disk image
    will be uploaded.

-   `image_name` (string) - The unique name of the resulting image.

-   `project_id` (string) - The project ID where the GCS bucket exists and
    where the GCE image is stored.

### Optional

-   `gcs_object_name` (string) - The name of the GCS object in `bucket` where
    the RAW disk image will be copied for import. Defaults to
    "packer-import-{{timestamp}}.tar.gz".

-   `image_description` (string) - The description of the resulting image.

-   `image_family` (string) - The name of the image family to which the
    resulting image belongs.

-   `image_labels` (object of key/value strings) - Key/value pair labels to
    apply to the created image.

-   `keep_input_artifact` (boolean) - if true, do not delete the compressed RAW
    disk image. Defaults to false.

-   `skip_clean` (boolean) - Skip removing the TAR file uploaded to the GCS
    bucket after the import process has completed. "true" means that we should
    leave it in the GCS bucket, "false" means to clean it out. Defaults to
    `false`.

## Basic Example

Here is a basic example. This assumes that the builder has produced an
compressed raw disk image artifact for us to work with, and that the GCS bucket
has been created.

``` json
{
  "type": "googlecompute-import",
  "account_file": "account.json",
  "project_id": "my-project",
  "bucket": "my-bucket",
  "image_name": "my-gce-image"
}
```

## QEMU Builder Example

Here is a complete example for building a Fedora 28 server GCE image. For this
example packer was run from a CentOS 7 server with KVM installed. The CentOS 7
server was running in GCE with the nested hypervisor feature enabled.

    $ packer build -var serial=$(tty) build.json

``` json
{
  "variables": {
    "serial": ""
  },
  "builders": [
    {
      "type": "qemu",
      "accelerator": "kvm",
      "communicator": "none",
      "boot_command": ["<tab> console=ttyS0,115200n8 inst.text inst.ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/fedora-28-ks.cfg rd.live.check=0<enter><wait>"],
      "disk_size": "15000",
      "format": "raw",
      "iso_checksum_type": "sha256",
      "iso_checksum": "ea1efdc692356b3346326f82e2f468903e8da59324fdee8b10eac4fea83f23fe",
      "iso_url": "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/28/Server/x86_64/iso/Fedora-Server-netinst-x86_64-28-1.1.iso",
      "headless": "true",
      "http_directory": "http",
      "http_port_max": "10089",
      "http_port_min": "10082",
      "output_directory": "output",
      "shutdown_timeout": "30m",
      "vm_name": "disk.raw",
      "qemu_binary": "/usr/libexec/qemu-kvm",
      "qemuargs": [
        [
          "-m", "1024"
        ],
        [
          "-cpu", "host"
        ],
        [
          "-chardev", "tty,id=pts,path={{user `serial`}}"
        ],
        [
          "-device", "isa-serial,chardev=pts"
        ],
        [
          "-device", "virtio-net,netdev=user.0"
        ]
      ]
    }
  ],
  "post-processors": [
    [
      {
        "type": "compress",
        "output": "output/disk.raw.tar.gz"
      },
      {
        "type": "googlecompute-import",
        "project_id": "my-project",
        "account_file": "account.json",
        "bucket": "my-bucket",
        "image_name": "fedora28-server-{{timestamp}}",
        "image_description": "Fedora 28 Server",
        "image_family": "fedora28-server"
      }
    ]
  ]
}
```
