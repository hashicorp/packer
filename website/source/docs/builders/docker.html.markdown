---
layout: "docs"
---

# Docker Builder

Type: `docker`

The Docker builder builds [Docker](http://www.docker.io) images using
Docker. The builder starts a Docker container, runs provisioners within
this container, then exports the container for re-use.

The Docker builder must run on a machine that supports Docker.

## Basic Example

Below is a fully functioning example. It doesn't do anything useful, since
no provisioners are defined, but it will effectively repackage an image.

<pre class="prettyprint">
{
  "type": "docker",
  "image": "ubuntu",
  "export_path": "image.tar"
}
</pre>

## Configuration Reference

All configuration options are currently required.

* `export_path` (string) - The path where the final container will be exported
  as a tar file.

* `image` (string) - The base image for the Docker container that will
  be started. This image will be pulled from the Docker registry if it
  doesn't already exist.

## Dockerfiles

TODO
