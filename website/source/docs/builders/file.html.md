---
description: |
    The `file` Packer builder is not really a builder, it just creates an artifact
    from a file. It can be used to debug post-processors without incurring high wait
    times. It does not run any provisioners.
layout: docs
page_title: File Builder
---

# File Builder

Type: `file`

The `file` Packer builder is not really a builder, it just creates an artifact
from a file. It can be used to debug post-processors without incurring high wait
times. It does not run any provisioners.

## Basic Example

Below is a fully functioning example. It doesn't do anything useful, since no
provisioners are defined, but it will connect to the specified host via ssh.

``` {.javascript}
{
  "type":         "file",
  "content":      "Lorem ipsum dolor sit amet",
  "target":       "dummy_artifact"
}
```

## Configuration Reference

Configuration options are organized below into two categories: required and
optional. Within each category, the available options are alphabetized and
described.

Any [communicator](/docs/templates/communicator.html) defined is ignored.

### Required:

-   `target` (string) - The path for a file which will be copied as
    the artifact.

### Optional:

You can only define one of `source` or `content`. If none of them is defined the
artifact will be empty.

-   `source` (string) - The path for a file which will be copied as
    the artifact.

-   `content` (string) - The content that will be put into the artifact.
