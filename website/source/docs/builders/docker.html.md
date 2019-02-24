---
description: |
    The docker Packer builder builds Docker images using Docker. The builder starts
    a Docker container, runs provisioners within this container, then exports the
    container for reuse or commits the image.
layout: docs
page_title: 'Docker - Builders'
sidebar_current: 'docs-builders-docker'
---

# Docker Builder

Type: `docker`

The `docker` Packer builder builds [Docker](https://www.docker.io) images using
Docker. The builder starts a Docker container, runs provisioners within this
container, then exports the container for reuse or commits the image.

Packer builds Docker containers *without* the use of
[Dockerfiles](https://docs.docker.com/engine/reference/builder/). By not using
`Dockerfiles`, Packer is able to provision containers with portable scripts or
configuration management systems that are not tied to Docker in any way. It
also has a simple mental model: you provision containers much the same way you
provision a normal virtualized or dedicated server. For more information, read
the section on [Dockerfiles](#dockerfiles).

The Docker builder must run on a machine that has Docker Engine installed.
Therefore the builder only works on machines that support Docker and *does not
support running on a Docker remote host*. You can learn about what [platforms
Docker supports and how to install onto
them](https://docs.docker.com/engine/installation/) in the Docker
documentation.

     Please note: Packer does not yet have support for Windows containers.

## Basic Example: Export

Below is a fully functioning example. It doesn't do anything useful, since no
provisioners are defined, but it will effectively repackage an image.

``` json
{
  "type": "docker",
  "image": "ubuntu",
  "export_path": "image.tar"
}
```

## Basic Example: Commit

Below is another example, the same as above but instead of exporting the
running container, this one commits the container to an image. The image can
then be more easily tagged, pushed, etc.

``` json
{
  "type": "docker",
  "image": "ubuntu",
  "commit": true
}
```

## Basic Example: Changes to Metadata

Below is an example using the changes argument of the builder. This feature
allows the source images metadata to be changed when committed back into the
Docker environment. It is derived from the `docker commit --change` command
line [option to
Docker](https://docs.docker.com/engine/reference/commandline/commit/).

Example uses of all of the options, assuming one is building an NGINX image
from ubuntu as an simple example:

``` json
{
  "type": "docker",
  "image": "ubuntu",
  "commit": true,
  "changes": [
    "USER www-data",
    "WORKDIR /var/www",
    "ENV HOSTNAME www.example.com",
    "VOLUME /test1 /test2",
    "EXPOSE 80 443",
    "LABEL version=1.0",
    "ONBUILD RUN date",
    "CMD [\"nginx\", \"-g\", \"daemon off;\"]",
    "ENTRYPOINT /var/www/start.sh"
  ]
}
```

Allowed metadata fields that can be changed are:

-   CMD
    -   String, supports both array (escaped) and string form
    -   EX: `"CMD [\"nginx\", \"-g\", \"daemon off;\"]"`
    -   EX: `"CMD nginx -g daemon off;"`
-   ENTRYPOINT
    -   String
    -   EX: `"ENTRYPOINT /var/www/start.sh"`
-   ENV
    -   String, note there is no equal sign:
    -   EX: `"ENV HOSTNAME www.example.com"` not
        `"ENV HOSTNAME=www.example.com"`
-   EXPOSE
    -   String, space separated ports
    -   EX: `"EXPOSE 80 443"`
-   LABEL
    -   String, space separated key=value pairs
    -   EX: `"LABEL version=1.0"`
-   ONBUILD
    -   String
    -   EX: `"ONBUILD RUN date"`
-   MAINTAINER
    -   String, deprecated in Docker version 1.13.0
    -   EX: `"MAINTAINER NAME"`
-   USER
    -   String
    -   EX: `"USER USERNAME"`
-   VOLUME
    -   String
    -   EX: `"VOLUME FROM TO"`
-   WORKDIR
    -   String
    -   EX: `"WORKDIR PATH"`

## Configuration Reference

Configuration options are organized below into two categories: required and
optional. Within each category, the available options are alphabetized and
described.

The Docker builder uses a special Docker communicator *and will not use* the
standard [communicators](/docs/templates/communicator.html).

### Required:

You must specify (only) one of `commit`, `discard`, or `export_path`.

-   `commit` (boolean) - If true, the container will be committed to an image
    rather than exported.

-   `discard` (boolean) - Throw away the container when the build is complete.
    This is useful for the [artifice
    post-processor](https://www.packer.io/docs/post-processors/artifice.html).

-   `export_path` (string) - The path where the final container will be
    exported as a tar file.

-   `image` (string) - The base image for the Docker container that will be
    started. This image will be pulled from the Docker registry if it doesn't
    already exist.

### Optional:

-   `author` (string) - Set the author (e-mail) of a commit.

-   `aws_access_key` (string) - The AWS access key used to communicate with
    AWS. [Learn how to set
    this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `aws_secret_key` (string) - The AWS secret key used to communicate with
    AWS. [Learn how to set
    this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `aws_token` (string) - The AWS access token to use. This is different from
    the access key and secret key. If you're not sure what this is, then you
    probably don't need it. This will also be read from the `AWS_SESSION_TOKEN`
    environmental variable.

-   `aws_profile` (string) - The AWS shared credentials profile used to
    communicate with AWS. [Learn how to set
    this.](/docs/builders/amazon.html#specifying-amazon-credentials)

-   `changes` (array of strings) - Dockerfile instructions to add to the
    commit. Example of instructions are `CMD`, `ENTRYPOINT`, `ENV`, and
    `EXPOSE`. Example: `[ "USER ubuntu", "WORKDIR /app", "EXPOSE 8080" ]`

-   `ecr_login` (boolean) - Defaults to false. If true, the builder will login
    in order to pull the image from [Amazon EC2 Container Registry
    (ECR)](https://aws.amazon.com/ecr/). The builder only logs in for the
    duration of the pull. If true `login_server` is required and `login`,
    `login_username`, and `login_password` will be ignored. For more
    information see the [section on ECR](#amazon-ec2-container-registry).

-   `exec_user` (string) - Username (UID) to run remote commands with. You can
    also set the group name/ID if you want: (UID or UID:GID).
    You may need this if you get permission errors trying to run the `shell` or
    other provisioners.

-   `login` (boolean) - Defaults to false. If true, the builder will login in
    order to pull the image. The builder only logs in for the duration of the
    pull. It always logs out afterwards. For log into ECR see `ecr_login`.

-   `login_username` (string) - The username to use to authenticate to login.

-   `login_password` (string) - The password to use to authenticate to login.

-   `login_server` (string) - The server address to login to.

-   `message` (string) - Set a message for the commit.

-   `privileged` (boolean) - If true, run the docker container with the
    `--privileged` flag. This defaults to false if not set.

-   `pull` (boolean) - If true, the configured image will be pulled using
    `docker pull` prior to use. Otherwise, it is assumed the image already
    exists and can be used. This defaults to true if not set.

-   `run_command` (array of strings) - An array of arguments to pass to
    `docker run` in order to run the container. By default this is set to
    `["-d", "-i", "-t", "{{.Image}}", "/bin/bash"]`. As you can see, you have a
    couple template variables to customize, as well.

-   `volumes` (map of strings to strings) - A mapping of additional volumes to
    mount into this container. The key of the object is the host path, the
    value is the container path.

-   `container_dir` (string) - The directory inside container to mount temp
    directory from host server for work [file
    provisioner](/docs/provisioners/file.html). By default this is set to
    `/packer-files`.

-   `fix_upload_owner` (boolean) - If true, files uploaded to the container
    will be owned by the user the container is running as. If false, the owner
    will depend on the version of docker installed in the system. Defaults to
    true.

## Using the Artifact: Export

Once the tar artifact has been generated, you will likely want to import, tag,
and push it to a container repository. Packer can do this for you automatically
with the [docker-import](/docs/post-processors/docker-import.html) and
[docker-push](/docs/post-processors/docker-push.html) post-processors.

**Note:** This section is covering how to use an artifact that has been
*exported*. More specifically, if you set `export_path` in your configuration.
If you set `commit`, see the next section.

The example below shows a full configuration that would import and push the
created image. This is accomplished using a sequence definition (a collection
of post-processors that are treated as as single pipeline, see
[Post-Processors](/docs/templates/post-processors.html) for more information):

``` json
{
  "post-processors": [
    [
      {
        "type": "docker-import",
        "repository": "hashicorp/packer",
        "tag": "0.7"
      },
      "docker-push"
    ]
  ]
}
```

In the above example, the result of each builder is passed through the defined
sequence of post-processors starting first with the `docker-import`
post-processor which will import the artifact as a docker image. The resulting
docker image is then passed on to the `docker-push` post-processor which
handles pushing the image to a container repository.

If you want to do this manually, however, perhaps from a script, you can import
the image using the process below:

``` shell
$ docker import - registry.mydomain.com/mycontainer:latest < artifact.tar
```

You can then add additional tags and push the image as usual with `docker tag`
and `docker push`, respectively.

## Using the Artifact: Committed

If you committed your container to an image, you probably want to tag, save,
push, etc. Packer can do this automatically for you. An example is shown below
which tags and pushes an image. This is accomplished using a sequence
definition (a collection of post-processors that are treated as as single
pipeline, see [Post-Processors](/docs/templates/post-processors.html) for more
information):

``` json
{
  "post-processors": [
    [
      {
        "type": "docker-tag",
        "repository": "hashicorp/packer",
        "tag": "0.7"
      },
      "docker-push"
    ]
  ]
}
```

In the above example, the result of each builder is passed through the defined
sequence of post-processors starting first with the `docker-tag` post-processor
which tags the committed image with the supplied repository and tag
information. Once tagged, the resulting artifact is then passed on to the
`docker-push` post-processor which handles pushing the image to a container
repository.

Going a step further, if you wanted to tag and push an image to multiple
container repositories, this could be accomplished by defining two,
nearly-identical sequence definitions, as demonstrated by the example below:

``` json
{
  "post-processors": [
    [
      {
        "type": "docker-tag",
        "repository": "hashicorp/packer1",
        "tag": "0.7"
      },
      "docker-push"
    ],
    [
      {
        "type": "docker-tag",
        "repository": "hashicorp/packer2",
        "tag": "0.7"
      },
      "docker-push"
    ]
  ]
}
```

<span id="amazon-ec2-container-registry"></span>

## Amazon EC2 Container Registry

Packer can tag and push images for use in [Amazon EC2 Container
Registry](https://aws.amazon.com/ecr/). The post processors work as described
above and example configuration properties are shown below:

``` json
{
  "post-processors": [
    [
      {
        "type": "docker-tag",
        "repository": "12345.dkr.ecr.us-east-1.amazonaws.com/packer",
        "tag": "0.7"
      },
      {
        "type": "docker-push",
        "ecr_login": true,
        "aws_access_key": "YOUR KEY HERE",
        "aws_secret_key": "YOUR SECRET KEY HERE",
        "login_server": "https://12345.dkr.ecr.us-east-1.amazonaws.com/"
      }
    ]
  ]
}
```

[Learn how to set Amazon AWS
credentials.](/docs/builders/amazon.html#specifying-amazon-credentials)

## Dockerfiles

This builder allows you to build Docker images *without* Dockerfiles.

With this builder, you can repeatedly create Docker images without the use of a
Dockerfile. You don't need to know the syntax or semantics of Dockerfiles.
Instead, you can just provide shell scripts, Chef recipes, Puppet manifests,
etc. to provision your Docker container just like you would a regular
virtualized or dedicated machine.

While Docker has many features, Packer views Docker simply as an container
runner. To that end, Packer is able to repeatedly build these containers using
portable provisioning scripts.

## Overriding the host directory

By default, Packer creates a temporary folder under your home directory, and
uses that to stage files for uploading into the container. If you would like to
change the path to this temporary folder, you can set the `PACKER_TMP_DIR`.
This can be useful, for example, if you have your home directory permissions
set up to disallow access from the docker daemon.
