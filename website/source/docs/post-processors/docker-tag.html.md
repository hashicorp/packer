---
description: |
    The Packer Docker Tag post-processor takes an artifact from the docker builder
    that was committed and tags it into a repository. This allows you to use the
    other Docker post-processors such as docker-push to push the image to a
    registry.
layout: docs
page_title: 'Docker Tag - Post-Processors'
sidebar_current: 'docs-post-processors-docker-tag'
---

# Docker Tag Post-Processor

Type: `docker-tag`

The Packer Docker Tag post-processor takes an artifact from the [docker
builder](/docs/builders/docker.html) that was committed and tags it into a
repository. This allows you to use the other Docker post-processors such as
[docker-push](/docs/post-processors/docker-push.html) to push the image to a
registry.

This is very similar to the
[docker-import](/docs/post-processors/docker-import.html) post-processor except
that this works with committed resources, rather than exported.

## Configuration

The configuration for this post-processor requires `repository`, all other settings
are optional.

-   `repository` (string) - The repository of the image.

-   `tag` (string) - The tag for the image. By default this is not set.

-   `force` (boolean) - If true, this post-processor forcibly tag the image even
    if tag name is collided. Default to `false`.
    But it will be ignored if Docker &gt;= 1.12.0 was detected,
    since the `force` option was removed after 1.12.0. [reference](https://docs.docker.com/engine/deprecated/#/f-flag-on-docker-tag)

## Example

An example is shown below, showing only the post-processor configuration:

``` json
{
  "type": "docker-tag",
  "repository": "hashicorp/packer",
  "tag": "0.7"
}
```

This example would take the image created by the Docker builder and tag it into
the local Docker process with a name of `hashicorp/packer:0.7`.

Following this, you can use the
[docker-push](/docs/post-processors/docker-push.html) post-processor to push it
to a registry, if you want.
