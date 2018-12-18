---
description: |
    The Packer Docker import post-processor takes an artifact from the docker
    builder and imports it with Docker locally. This allows you to apply a
    repository and tag to the image and lets you use the other Docker
    post-processors such as docker-push to push the image to a registry.
layout: docs
page_title: 'Docker Import - Post-Processors'
sidebar_current: 'docs-post-processors-docker-import'
---

# Docker Import Post-Processor

Type: `docker-import`

The Packer Docker import post-processor takes an artifact from the [docker
builder](/docs/builders/docker.html) and imports it with Docker locally. This
allows you to apply a repository and tag to the image and lets you use the
other Docker post-processors such as
[docker-push](/docs/post-processors/docker-push.html) to push the image to a
registry.

## Configuration

The configuration for this post-processor only requires a `repository`, a `tag`
is optional.

### Required:

-   `repository` (string) - The repository of the imported image.

-   `tag` (string) - The tag for the imported image. By default this is not
    set.

### Optional:

-   `changes` (array of strings) - Dockerfile instructions to add to the
    commit. Example of instructions are `CMD`, `ENTRYPOINT`, `ENV`, and
    `EXPOSE`. Example: `[ "USER ubuntu", "WORKDIR /app", "EXPOSE 8080" ]`


## Example

An example is shown below, showing only the post-processor configuration:

``` json
{
  "type": "docker-import",
  "repository": "hashicorp/packer",
  "tag": "0.7"
}
```

This example would take the image created by the Docker builder and import it
into the local Docker process with a name of `hashicorp/packer:0.7`.

Following this, you can use the
[docker-push](/docs/post-processors/docker-push.html) post-processor to push it
to a registry, if you want.

## Changing Metadata

Below is an example using the changes argument of the post-processor. This
feature allows the tarball metadata to be changed when imported into the
Docker environment. It is derived from the `docker import --change` command
line [option to
Docker](https://docs.docker.com/engine/reference/commandline/import/).

Example uses of all of the options, assuming one is building an NGINX image
from ubuntu as an simple example:

``` json
{
  "type": "docker-import",
  "repository": "local/centos6",
  "tag": "latest",
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
