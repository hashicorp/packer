---
layout: "docs"
page_title: "compress Post-Processor"
description: |-
  The Packer compress post-processor takes an artifact with files (such as from VMware or VirtualBox) and compresses the artifact into a single archive.
---

# Compress Post-Processor

Type: `compress`

The Packer compress post-processor takes an artifact with files (such as from
VMware or VirtualBox) and compresses the artifact into a single archive.

## Configuration

### Required:

You must specify the output filename. The archive format is derived from the filename.

* `output` (string) - The path to save the compressed archive. The archive
  format is inferred from the filename. E.g. `.tar.gz` will be a gzipped
  tarball. `.zip` will be a zip file. If the extension can't be detected packer
  defaults to `.tar.gz` behavior but will not change the filename.

  If you are executing multiple builders in parallel you should make sure
  `output` is unique for each one. For example `packer_{{.BuildName}}_{{.Provider}}.zip`.

### Optional:

If you want more control over how the archive is created you can specify the following settings:

* `compression_level` (integer) - Specify the compression level, for algorithms
  that support it, from 1 through 9 inclusive. Typically higher compression
  levels take longer but produce smaller files. Defaults to `6`

* `keep_input_artifact` (bool) - Keep source files; defaults to `false`

### Supported Formats

Supported file extensions include `.zip`, `.tar`, `.gz`, `.tar.gz`, `.lz4` and `.tar.lz4`. Note that `.gz` and `.lz4` will fail if you have multiple files to compress.

## Examples

Some minimal examples are shown below, showing only the post-processor configuration:

```json
{
  "type": "compress",
  "output": "archive.tar.lz4"
}
```

```json
{
  "type": "compress",
  "output": "archive.zip"
}
```

```json
{
  "type": "compress",
  "output": "archive.gz",
  "compression": 9
}
```
