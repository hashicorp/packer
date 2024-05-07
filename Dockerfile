# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: BUSL-1.1

# ========================================================================
#
# This Dockerfile contains multiple targets.
# Use 'docker build --target=<name> .' to build one.
# e.g. `docker build --target=release-light .`
#
# All non-dev targets have a PRODUCT_VERSION argument that must be provided
# via --build-arg=PRODUCT_VERSION=<version> when building.
# e.g. --build-arg PRODUCT_VERSION=1.11.2
#
# For local dev and testing purposes, please build and use the `dev` docker image.
#
# ========================================================================


# Development docker image primarily used for development and debugging.
# This image builds from the locally generated binary in ./bin/.
# To generate the local binary, run `make dev`.
FROM docker.mirror.hashicorp.services/alpine:latest as dev

RUN apk add --no-cache git bash openssl ca-certificates

COPY bin/packer /bin/packer

ENTRYPOINT ["/bin/packer"]

# Light docker image which can be used to run the binary from a container.
# This image builds from the locally generated binary in ./bin/, and from CI-built binaries within CI.
# To generate the local binary, run `make dev`.
# This image is published to DockerHub under the `light`, `light-$VERSION`, and `latest` tags.
FROM docker.mirror.hashicorp.services/alpine:latest as release-light

ARG PRODUCT_VERSION
ARG BIN_NAME

# TARGETARCH and TARGETOS are set automatically when --platform is provided.
ARG TARGETOS TARGETARCH

LABEL name="Packer" \
      maintainer="HashiCorp Packer Team <packer@hashicorp.com>" \
      vendor="HashiCorp" \
      version=$PRODUCT_VERSION \
      release=$PRODUCT_VERSION \
      summary="Packer is a tool for creating identical machine images for multiple platforms from a single source configuration." \
      description="Packer is a tool for creating identical machine images for multiple platforms from a single source configuration. Please submit issues to https://github.com/hashicorp/packer/issues" \
      org.opencontainers.image.licenses="BUSL-1.1"

RUN apk add --no-cache git bash wget openssl gnupg xorriso

COPY dist/$TARGETOS/$TARGETARCH/$BIN_NAME /bin/
RUN mkdir -p /usr/share/doc/Packer
COPY LICENSE /usr/share/doc/Packer/LICENSE.txt

ENTRYPOINT ["/bin/packer"]

# Full docker image which can be used to run the binary from a container.
# This image is essentially the same as the `release-light` one, but embeds
# the official plugins in it.
FROM release-light as release-full

# Install the latest version of the official plugins
RUN /bin/packer plugins install "github.com/hashicorp/amazon" && \
    /bin/packer plugins install "github.com/hashicorp/ansible" && \
    /bin/packer plugins install "github.com/hashicorp/azure" && \
    /bin/packer plugins install "github.com/hashicorp/docker" && \
    /bin/packer plugins install "github.com/hashicorp/googlecompute" && \
    /bin/packer plugins install "github.com/hashicorp/qemu" && \
    /bin/packer plugins install "github.com/hashicorp/vagrant" && \
    /bin/packer plugins install "github.com/hashicorp/virtualbox" && \
    /bin/packer plugins install "github.com/hashicorp/vmware" && \
    /bin/packer plugins install "github.com/hashicorp/vsphere"

ENTRYPOINT ["/bin/packer"]

# Set default target to 'dev'.
FROM dev
