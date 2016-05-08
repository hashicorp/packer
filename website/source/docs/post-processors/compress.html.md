---
description: |
    The Packer compress post-processor takes an artifact with files (such as from
    VMware or VirtualBox) and compresses the artifact into a single archive.
layout: docs
page_title: 'compress Post-Processor'
...

# Compress Post-Processor

Type: `compress`

The Packer compress post-processor takes an artifact with files (such as from
VMware or VirtualBox) and compresses the artifact into a single archive.

## Configuration

### Optional:

By default, packer will build archives in `.tar.gz` format with the following
filename: `packer_{{.BuildName}}_{{.BuilderType}}`. If you want to change this
you will need to specify the `output` option.

-   `output` (string) - The path to save the compressed archive. The archive
    format is inferred from the filename. E.g. `.tar.gz` will be a
    gzipped tarball. `.zip` will be a zip file. If the extension can't be
    detected packer defaults to `.tar.gz` behavior but will not change
    the filename.

    You can use `{{.BuildName}}` and `{{.BuilderType}}` in your output path. If
    you are executing multiple builders in parallel you should make sure
    `output` is unique for each one. For example `packer_{{.BuildName}}.zip`.

-   `format` (string) - Disable archive format autodetection and use provided
    string.

-   `compression_level` (integer) - Specify the compression level, for
    algorithms that support it, from 1 through 9 inclusive. Typically higher
    compression levels take longer but produce smaller files. Defaults to `6`

-   `keep_input_artifact` (boolean) - Keep source files; defaults to `false`

### Supported Formats

Supported file extensions include `.zip`, `.tar`, `.gz`, `.tar.gz`, `.lz4` and
`.tar.lz4`. Note that `.gz` and `.lz4` will fail if you have multiple files to
compress.

## Examples

Some minimal examples are shown below, showing only the post-processor
configuration:

``` {.json}
{
  "type": "compress",
  "output": "archive.tar.lz4"
}
```

``` {.json}
{
  "type": "compress",
  "output": "{{.BuildName}}_bundle.zip"
}
```

``` {.json}
{
  "type": "compress",
  "output": "log_{{.BuildName}}.gz",
  "compression_level": 9
}
```
