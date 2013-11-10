---
layout: "docs"
---

# Docker Builder

Type: `docker`

The Docker builder builds [Docker](http://www.docker.io) images using
Docker. The builder starts a Docker container, runs provisioners within
this container, then exports the container for re-use.

Packer builds Docker containers _without_ the use of
[Dockerfiles](http://docs.docker.io/en/latest/use/builder/).
By not using Dockerfiles, Packer is able to provision
containers with portable scripts or configuration management systems
that are not tied to Docker in any way. It also has a simpler mental model:
you provision containers much the same way you provision a normal virtualized
or dedicated server. For more information, read the section on
[Dockerfiles](#toc_3).

The Docker builder must run on a machine that has Docker installed. Therefore
the builder only works on machines that support Docker (modern Linux machines).
If you want to use Packer to build Docker containers on another platform,
use [Vagrant](http://www.vagrantup.com) to start a Linux environment, then
run Packer within that environment.

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

Configuration options are organized below into two categories: required and optional. Within
each category, the available options are alphabetized and described.

Required:

* `export_path` (string) - The path where the final container will be exported
  as a tar file.

* `image` (string) - The base image for the Docker container that will
  be started. This image will be pulled from the Docker registry if it
  doesn't already exist.

Optional:

* `pull` (bool) - If true, the configured image will be pulled using
  `docker pull` prior to use. Otherwise, it is assumed the image already
  exists and can be used. This defaults to true if not set.

## Dockerfiles

This builder allows you to build Docker images _without_ Dockerfiles.

With this builder, you can repeatably create Docker images without the use
a Dockerfile. You don't need to know the syntax or semantics of Dockerfiles.
Instead, you can just provide shell scripts, Chef recipes, Puppet manifests,
etc. to provision your Docker container just like you would a regular
virtualized or dedicated machine.

While Docker has many features, Packer views Docker simply as an LXC
container runner. To that end, Packer is able to repeatably build these
LXC containers using portable provisioning scripts.

Dockerfiles have some additional features that Packer doesn't support
which are able to be worked around. Many of these features will be automated
by Packer in the future:

* Dockerfiles will snapshot the container at each step, allowing you to
  go back to any step in the history of building. Packer doesn't do this yet,
  but inter-step snapshotting is on the way.

* Dockerfiles can contain information such as exposed ports, shared
  volumes, and other metadata. Packer builds a raw Docker container image
  that has none of this metadata. You can pass in much of this metadata
  at runtime with `docker run`.

* Images made without dockerfiles are missing critical metadata that
  make them easily pushable to the Docker registry. You can work around
  this by using a metadata-only Dockerfile with the exported image and
  building that. A future Packer version will automatically do this for you.
