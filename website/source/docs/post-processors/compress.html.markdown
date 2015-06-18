---
layout: "docs"
page_title: "compress Post-Processor"
description: |-
  The Packer compress post-processor takes an artifact with files (such as from VMware or VirtualBox) and gzip compresses the artifact into a single archive.
---

# Compress Post-Processor

Type: `compress`

The Packer compress post-processor takes an artifact with files (such as from
VMware or VirtualBox) and gzip compresses the artifact into a single
archive.

## Configuration

The minimal required configuration is to specify the output file. This will create a gzipped tarball.

* `output` (required, string) - The path to save the compressed archive. The archive format is inferred from the filename. E.g. `.tar.gz` will be a gzipped tarball. `.zip` will be a zip file.

  If the extension can't be detected tar+gzip will be used as a fallback.

If you want more control over how the archive is created you can specify the following settings:

* `level` (optional, integer) - Specify the compression level, for algorithms that support it. Value from -1 through 9 inclusive. 9 offers the smallest file size, but takes longer
* `keep_input_artifact` (optional, bool) - Keep source files; defaults to false

## Supported Formats

Supported file extensions include `.zip`, `.tar`, `.gz`, `.tar.gz`, `.lz4` and `.tar.lz4`.

## Example

Some minimal examples are shown below, showing only the post-processor configuration:

```json
{
  "type": "compress",
  "output": "archive.tar.gz"
}
```

```json
{
  "type": "compress",
  "output": "archive.zip"
}
```

A more complex example, again showing only the post-processor configuration:

```json
{
  "type": "compress",
  "output": "archive.gz",
  "compression": 9,
  "parallel": false
}
```
