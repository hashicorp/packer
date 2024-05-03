# Packer Docker Container

The root of this repository contains the officially supported HashiCorp Dockerfile to build the hashicorp/packer docker image. The `dev` docker image should be built for local dev and testing, while the production docker image, `release`, is built in CI and makes use of CI-built binaries. The `light` and `full` docker images are built using the official binaries from releases.hashicorp.com.

## License

This image is licensed under (BUSL-1.1)[https://github.com/hashicorp/packer/blob/main/LICENSE].

## Build

Refer to the Makefile of this repository, especially the `docker` and `docker-dev` targets to build a local version of the dev image based on the sources available.

### Usage

This repository automatically builds containers for using the
[`packer`](https://developer.hashicorp.com/packer) command line program. It contains three distinct
varieties of build: a `light` version, which just contains the binary,
a `full` build, which contains the Packer binary with pre-installed plugins,
and a `dev` version, which compiles the binary from source
inside the container before exposing it for use.

##### `light`

The `light` version of this container will copy the current stable version of
the binary, taken from releases.hashicorp.com, into the container. It will also
set it for use as the default entrypoint. This will be the best option for most uses,
especially if you are just looking to run the binary from a container.
The `latest` tag on DockerHub also points to this version.

You can use this version with the following:
```shell
docker run <args> hashicorp/packer:light <command>
```

##### `full`

The `full` version of the container builds upon `light` and pre-installs
the plugins officially maintained by HashiCorp.

You can use this version with the following:
```shell
docker run <args> hashicorp/packer:full <command>
```

You can view the list of pre-installed plugins with the following:
```shell
docker run <args> hashicorp/packer:full plugins installed
```

##### `dev`

The `dev` version of this container contains all of the source code found in
the current ref of this [repository](https://github.com/hashicorp/packer). Using [Google's
official `golang` image](https://hub.docker.com/_/golang/) as a base, this
container will copy the source from the current branch, build the binary, and
expose it for running. Because all build artifacts are included, it should be quite a bit larger than
the `light` image. This version of the container is most useful for development or
debugging.

You can use this version with the following:
```shell
docker run <args> hashicorp/packer:dev <command>
```

#### Running a build:

The easiest way to run a command that references a configuration with one or more template files, is to mount a volume for the local workspace.

Running `packer init`
```shell
docker run \
    -v `pwd`:/workspace -w /workspace \
    -e PACKER_PLUGIN_PATH=/workspace/.packer.d/plugins \
    hashicorp/packer:latest \
    init .
```

~> **Note**: packer init is available from Packer v1.7.0 and later

The command will mount the working directory (`pwd`) to `workspace`, which is the working directory (`-w`) inside the container.
Any plugin installed with `packer init` will be installed under the directory specified under the `PACKER_PLUGIN_PATH` environment variable. `PACKER_PLUGIN_PATH` must be set to a path inside the volume mount so that plugins can become available at `packer build`.

Running `packer build`
```shell
docker run \
    -v `pwd`:/workspace -w /workspace \
    -e PACKER_PLUGIN_PATH=/workspace/.packer.d/plugins \
    hashicorp/packer:latest \
    build .
```
##### Building old-legacy JSON templates

For old-legacy JSON, the build command must specify the template file(s).

```shell
docker run \
    -v `pwd`:/workspace -w /workspace \
    hashicorp/packer:latest \
    build template.json
```

For the [manual installation](https://www.packer.io/docs/plugins#installing-plugins) of third-party plugins, we recommended that plugin binaries are placed under a sub-directory under the working directory. Add `-e PACKER_PLUGIN_PATH=/workspace/<subdirectory_plugin_path>` to the command above to tell Packer where the plugins are.

To pass a var file (`var.json`) to the build command:

```shell
docker run \
    -v `pwd`:/workspace -w /workspace \
    hashicorp/packer:latest \
    build --var-file var.json template.json
```
`var.json` is expected to be inside the local working directory (`pwd`) and in the container's workspace mount.
