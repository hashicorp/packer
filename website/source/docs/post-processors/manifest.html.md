---
description: |
    The manifest post-processor writes a JSON file with the build artifacts and IDs from a packer run.
layout: docs
page_title: 'Manifest Post-Processor'
...

# Manifest Post-Processor

Type: `manifest`

The manifest post-processor writes a JSON file with a list of all of the artifacts packer produces during a run. If your packer template includes multiple builds, this helps you keep track of which output artifacts (files, AMI IDs, docker containers, etc.) correspond to each build.

The manifest post-processor is invoked each time a build completes and *updates* data in the manifest file. Builds are identified by name and type, and include their build time, artifact ID, and file list.

If packer is run with the `-force` flag the manifest file will be truncated automatically during each packer run. Otherwise, subsequent builds will be added to the file. You can use the timestamps to see which is the latest artifact.

You can specify manifest more than once and write each build to its own file, or write all builds to the same file. For simple builds manifest only needs to be specified once (see below) but you can also chain it together with other post-processors such as Docker and Artifice.

## Configuration

### Optional:

-   `filename` (string) The manifest will be written to this file. This defaults to `packer-manifest.json`.

### Example Configuration

You can simply add `{"type":"manifest"}` to your post-processor section. Below is a more verbose example:

``` {.javascript}
{
  "post-processors": [
    {
      "type": "manifest",
      "filename": "manifest.json"
    }
  ]
}
```
