---
description: |
    The manifest post-processor writes a JSON file with the build artifacts and
    IDs from a packer run.
layout: docs
page_title: 'Manifest - Post-Processors'
sidebar_current: 'docs-post-processors-manifest'
---

# Manifest Post-Processor

Type: `manifest`

The manifest post-processor writes a JSON file with a list of all of the artifacts packer produces during a run. If your packer template includes multiple builds, this helps you keep track of which output artifacts (files, AMI IDs, docker containers, etc.) correspond to each build.

The manifest post-processor is invoked each time a build completes and *updates* data in the manifest file. Builds are identified by name and type, and include their build time, artifact ID, and file list.

If packer is run with the `-force` flag the manifest file will be truncated automatically during each packer run. Otherwise, subsequent builds will be added to the file. You can use the timestamps to see which is the latest artifact.

You can specify manifest more than once and write each build to its own file, or write all builds to the same file. For simple builds manifest only needs to be specified once (see below) but you can also chain it together with other post-processors such as Docker and Artifice.

## Configuration

### Optional:

-   `output` (string) The manifest will be written to this file. This defaults to `packer-manifest.json`.
-   `strip_path` (boolean) Write only filename without the path to the manifest file. This defaults to false.

### Example Configuration

You can simply add `{"type":"manifest"}` to your post-processor section. Below is a more verbose example:

``` json
{
  "post-processors": [
    {
      "type": "manifest",
      "output": "manifest.json",
      "strip_path": true
    }
  ]
}
```

An example manifest file looks like:

``` json
{
  "builds": [
    {
      "name": "docker",
      "builder_type": "docker",
      "build_time": 1507245986,
      "files": [
        {
          "name": "packer_example",
          "size": 102219776
        }
      ],
      "artifact_id": "Container",
      "packer_run_uuid": "6d5d3185-fa95-44e1-8775-9e64fe2e2d8f"
    }
  ],
  "last_run_uuid": "6d5d3185-fa95-44e1-8775-9e64fe2e2d8f"
}
```

If the build is run again, the new build artifacts will be added to the manifest file rather than replacing it. It is possible to grab specific build artifacts from the manifest by using `packer_run_uuid`.

The above mainfest was generated with this packer.json:

```json
{
  "builders": [
    {
      "type":        "docker",
      "image":       "ubuntu:latest",
      "export_path": "packer_example",
      "run_command": [ "-d", "-i", "-t", "--entrypoint=/bin/bash", "{{.Image}}" ]
    }
  ],
    "provisioners": [
    {
      "type": "shell",
      "inline": "mkdir /Setup"
    },
    {
      "type": "file",
      "source": "../scripts/dummy_bash.sh",
      "destination": "/Setup"
    },
    {
      "type": "shell",
      "inline":["ls -alh /Setup/"]
    }
  ],
  "post-processors": [
    {
      "type": "manifest",
      "output": "manifest.json",
      "strip_path": true
    }
  ]
}
```
