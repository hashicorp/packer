---
description: |
    The Packer Docker import post-processor takes an artifact from the docker
    builder and imports it with Docker locally. This allows you to apply a
    repository and tag to the image and lets you use the other Docker
    post-processors such as docker-push to push the image to a registry.
layout: docs
page_title: 'docker-import Post-Processor'
...

# Docker Import Post-Processor

Type: `docker-import`

The Packer Docker import post-processor takes an artifact from the [docker
builder](/docs/builders/docker.html) and imports it with Docker locally. This
allows you to apply a repository and tag to the image and lets you use the other
Docker post-processors such as
[docker-push](/docs/post-processors/docker-push.html) to push the image to a
registry.

## Configuration

The configuration for this post-processor is extremely simple. At least a
repository is required.

-   `repository` (string) - The repository of the imported image.

-   `tag` (string) - The tag for the imported image. By default this is not set.

## Example

An example is shown below, showing only the post-processor configuration:

``` {.javascript}
{
  "type": "docker-import",
  "repository": "mitchellh/packer",
  "tag": "0.7"
}
```

This example would take the image created by the Docker builder and import it
into the local Docker process with a name of `mitchellh/packer:0.7`.

Following this, you can use the
[docker-push](/docs/post-processors/docker-push.html) post-processor to push it
to a registry, if you want.
