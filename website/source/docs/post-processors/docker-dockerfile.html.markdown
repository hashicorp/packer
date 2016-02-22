---
layout: "docs"
page_title: "docker-dockerfile Post-Processor"
description: |-
  The Packer Docker Dockerfile post-processor takes an artifact from the docker builder that was committed and adds information such as exposed ports, shared volumes, and other metadata by using Dockerfile. This allows you to use the other Docker post-processors such as docker-tag to tag it or docker-push to push the image to a registry.
---

# Docker Dockerfile Post-Processor

Type: `docker-dockerfile`

The Packer Docker Dockerfile post-processor takes an artifact from the
[docker builder](/docs/builders/docker.html) that was committed
and adds information such as exposed ports, shared volumes, and other
metadata by using [Dockerfile](https://docs.docker.com/reference/builder/).
This allows you to use the other Docker post-processors such as
[docker-tag](/docs/post-processors/docker-tag.html) to tag it or
[docker-push](/docs/post-processors/docker-push.html) to push the image
to a registry.

This is very similar to the
[docker-import](/docs/post-processors/docker-import.html) post-processor
except that this works with committed resources, rather than exported.

## Configuration

The configuration options for this post-processor are similar to Dockerfile
instructions. Every option is optional. More information can be found in the
[Dockerfile Reference](https://docs.docker.com/reference/builder/). `RUN`,
`ADD`, `COPY`, and `ONBUILD` Dockerfile instructions are not supported because
Packer provisioners should be used for their functionality.

* `maintainer` (string) - The author field of the generated images.

* `cmd` (array of strings) - An array of a command and/or parameters as
  defaults for an executing container.

* `label` (map of strings to strings) - Key-value pairs of additional metadata.

* `expose` (array of strings) - An array of network ports that the container
  will expose to the host.

* `env` (map of strings to strings) - Key-value pairs of environment variables
  for all "descendent" images.

* `entrypoint` (array of strings) - An array of a command and its parameters to
  configure a container that will run as an executable.

* `volume` (array of strings) - An array of container paths to create mount
  points for externally mounted voluems from native host or other containers.

* `user` (string) - The username or UID to use when running the image and for
  any `CMD` and `ENTRYPOINT` instructions.

* `workdir` (string) - The working directory for any `CMD` and `ENTRYPOINT`
  instructions.

## Example

An example is shown below, showing only the post-processor configuration:

```javascript
{
  "type": "docker-dockerfile",
  "maintainer": "James G. Kim <jgkim@jayg.org>",
  "cmd": ["/bin/bash"],
  "label": {
    "version": "1.0"
  },
  "expose": [8080],
  "volume": ["/var/log"]
}
```

This example would take the image created by the Docker builder and add
metadata to it by using a Dockerfile.

Following this, you can use the
[docker-tag](/docs/post-processors/docker-tag.html) post-processor to tag
it or the [docker-push](/docs/post-processors/docker-push.html) post-processor
to push it to a registry, if you want.
