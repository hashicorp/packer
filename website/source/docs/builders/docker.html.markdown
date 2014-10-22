---
layout: "docs"
page_title: "Docker Builder"
---

# Docker Builder

Type: `docker`

The Docker builder builds [Docker](http://www.docker.io) images using
Docker. The builder starts a Docker container, runs provisioners within
this container, then exports the container for reuse or commits the image.

Packer builds Docker containers _without_ the use of
[Dockerfiles](http://docs.docker.io/en/latest/use/builder/).
By not using Dockerfiles, Packer is able to provision
containers with portable scripts or configuration management systems
that are not tied to Docker in any way. It also has a simpler mental model:
you provision containers much the same way you provision a normal virtualized
or dedicated server. For more information, read the section on
[Dockerfiles](#toc_4).

The Docker builder must run on a machine that has Docker installed. Therefore
the builder only works on machines that support Docker (modern Linux machines).
If you want to use Packer to build Docker containers on another platform,
use [Vagrant](http://www.vagrantup.com) to start a Linux environment, then
run Packer within that environment.

## Basic Example: Export

Below is a fully functioning example. It doesn't do anything useful, since
no provisioners are defined, but it will effectively repackage an image.

```javascript
{
  "type": "docker",
  "image": "ubuntu",
  "export_path": "image.tar"
}
```

## Basic Example: Commit

Below is another example, the same as above but instead of exporting the
running container, this one commits the container to an image. The image
can then be more easily tagged, pushed, etc.

```javascript
{
  "type": "docker",
  "image": "ubuntu",
  "commit": true
}
```


## Configuration Reference

Configuration options are organized below into two categories: required and
optional. Within each category, the available options are alphabetized and
described.

### Required:

* `commit` (boolean) - If true, the container will be committed to an
  image rather than exported. This cannot be set if `export_path` is set.

* `export_path` (string) - The path where the final container will be exported
  as a tar file. This cannot be set if `commit` is set to true.

* `image` (string) - The base image for the Docker container that will
  be started. This image will be pulled from the Docker registry if it
  doesn't already exist.

### Optional:

* `login` (boolean) - Defaults to false. If true, the builder will
    login in order to pull the image. The builder only logs in for the
    duration of the pull. It always logs out afterwards.

* `login_email` (string) - The email to use to authenticate to login.

* `login_username` (string) - The username to use to authenticate to login.

* `login_password` (string) - The password to use to authenticate to login.

* `login_server` (string) - The server address to login to.

* `pull` (boolean) - If true, the configured image will be pulled using
  `docker pull` prior to use. Otherwise, it is assumed the image already
  exists and can be used. This defaults to true if not set.

* `run_command` (array of strings) - An array of arguments to pass to
  `docker run` in order to run the container. By default this is set to
  `["-d", "-i", "-t", "{{.Image}}", "/bin/bash"]`.
  As you can see, you have a couple template variables to customize, as well.

* `volumes` (map of strings to strings) - A mapping of additional volumes
   to mount into this container. The key of the object is the host path,
   the value is the container path.

## Using the Artifact: Export

Once the tar artifact has been generated, you will likely want to import, tag,
and push it to a container repository. Packer can do this for you automatically
with the [docker-import](/docs/post-processors/docker-import.html) and
[docker-push](/docs/post-processors/docker-push.html) post-processors.

**Note:** This section is covering how to use an artifact that has been
_exported_. More specifically, if you set `export_path` in your configuration.
If you set `commit`, see the next section.

The example below shows a full configuration that would import and push
the created image:

```javascript
{
  "post-processors": [
		[
			{
				"type": "docker-import",
				"repository": "mitchellh/packer",
				"tag": "0.7"
			},
			"docker-push"
		]
	]
}
```

If you want to do this manually, however, perhaps from a script, you can
import the image using the process below:

```text
$ docker import - registry.mydomain.com/mycontainer:latest < artifact.tar
```

You can then add additional tags and push the image as usual with `docker tag`
and `docker push`, respectively.

## Using the Artifact: Committed

If you committed your container to an image, you probably want to tag,
save, push, etc. Packer can do this automatically for you. An example is
shown below which tags and pushes the image:

```javascript
{
  "post-processors": [
		[
			{
				"type": "docker-tag",
				"repository": "mitchellh/packer",
				"tag": "0.7"
			},
			"docker-push"
		]
	]
}
```

## Dockerfiles

This builder allows you to build Docker images _without_ Dockerfiles.

With this builder, you can repeatably create Docker images without the use of
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
