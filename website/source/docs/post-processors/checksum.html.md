---
description: |
    The checksum post-processor computes specified checksum for the artifact list
    from an upstream builder or post-processor. All downstream post-processors
    will see the new artifacts. The primary use-case is compute checksum for
    artifacts allows to verify it later. So firstly this post-processor get
    artifact, compute it checksum and pass to next post-processor original
    artifacts and checksum files.
layout: docs
page_title: 'Checksum - Post-Processors'
sidebar_current: 'docs-post-processors-checksum'
---

# Checksum Post-Processor

Type: `checksum`

The checksum post-processor computes specified checksum for the artifact list
from an upstream builder or post-processor. All downstream post-processors will
see the new artifacts. The primary use-case is compute checksum for artifact to
verify it later.

After computes checksum for artifacts, you can use new artifacts with other
post-processors like
[artifice](https://www.packer.io/docs/post-processors/artifice.html),
[compress](https://www.packer.io/docs/post-processors/compress.html),
[docker-push](https://www.packer.io/docs/post-processors/docker-push.html),
[atlas](https://www.packer.io/docs/post-processors/atlas.html), or a third-party
post-processor.

## Basic example

The example below is fully functional.

``` json
{
  "type": "checksum"
}
```

## Configuration Reference

Optional parameters:

-   `checksum_types` (array of strings) - An array of strings of checksum types
    to compute. Allowed values are md5, sha1, sha224, sha256, sha384, sha512.
-   `output` (string) - Specify filename to store checksums. This defaults to
    `packer_{{.BuildName}}_{{.BuilderType}}_{{.ChecksumType}}.checksum`. For
    example, if you had a builder named `database`, you might see the file
    written as `packer_database_docker_md5.checksum`. The following variables are
    available to use in the output template:

    -   `BuildName`: The name of the builder that produced the artifact.
    -   `BuilderType`: The type of builder used to produce the artifact.
    -   `ChecksumType`: The type of checksums the file contains. This should be
        used if you have more than one value in `checksum_types`.
