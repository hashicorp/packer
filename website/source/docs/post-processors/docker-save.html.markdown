---
description: |
    The Packer Docker Save post-processor takes an artifact from the docker builder
    that was committed and saves it to a file. This is similar to exporting the
    Docker image directly from the builder, except that it preserves the hierarchy
    of images and metadata.
layout: docs
page_title: 'docker-save Post-Processor'
...

# Docker Save Post-Processor

Type: `docker-save`

The Packer Docker Save post-processor takes an artifact from the [docker
builder](/docs/builders/docker.html) that was committed and saves it to a file.
This is similar to exporting the Docker image directly from the builder, except
that it preserves the hierarchy of images and metadata.

We understand the terminology can be a bit confusing, but we've adopted the
terminology from Docker, so if you're familiar with that, then you'll be
familiar with this and vice versa.

## Configuration

The configuration for this post-processor is extremely simple.

-   `path` (string) - The path to save the image.

## Example

An example is shown below, showing only the post-processor configuration:

``` {.javascript}
{
  "type": "docker-save",
  "path": "foo.tar"
}
```
