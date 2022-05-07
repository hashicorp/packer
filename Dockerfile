# ========================================================================
# 
# This Dockerfile contains multiple targets.
# Use 'docker build --target=<name> .' to build one.
# e.g. `docker build --target=release-light .`
#
# All non-dev targets have a VERSION argument that must be provided 
# via --build-arg=VERSION=<version> when building. 
# e.g. --build-arg VERSION=1.11.2
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


# Official docker image that includes binaries from releases.hashicorp.com. 
# This downloads the release from releases.hashicorp.com and therefore requires that
# the release is published before building the Docker image.
FROM docker.mirror.hashicorp.services/alpine:latest as official

# This is the release of Packer to pull in.
ARG VERSION

LABEL name="Packer" \
      maintainer="HashiCorp Packer Team <packer@hashicorp.com>" \
      vendor="HashiCorp" \
      version=$VERSION \
      release=$VERSION \
      summary="Packer is a tool for creating identical machine images for multiple platforms from a single source configuration." \
      description="Packer is a tool for creating identical machine images for multiple platforms from a single source configuration. Please submit issues to https://github.com/hashicorp/packer/issues"

# This is the location of the releases.
ENV HASHICORP_RELEASES=https://releases.hashicorp.com

RUN set -eux && \
    apk add --no-cache git bash wget openssl gnupg && \
    gpg --keyserver keyserver.ubuntu.com --recv-keys C874011F0AB405110D02105534365D9472D7468F && \
    mkdir -p /tmp/build && \
    cd /tmp/build && \
    apkArch="$(apk --print-arch)" && \
    case "${apkArch}" in \
        aarch64) packerArch='arm64' ;; \
        armhf) packerArch='arm' ;; \
        x86) packerArch='386' ;; \
        x86_64) packerArch='amd64' ;; \
        *) echo >&2 "error: unsupported architecture: ${apkArch} (see ${HASHICORP_RELEASES}/packer/${VERSION}/)" && exit 1 ;; \
    esac && \
    wget ${HASHICORP_RELEASES}/packer/${VERSION}/packer_${VERSION}_linux_${packerArch}.zip && \
    wget ${HASHICORP_RELEASES}/packer/${VERSION}/packer_${VERSION}_SHA256SUMS && \
    wget ${HASHICORP_RELEASES}/packer/${VERSION}/packer_${VERSION}_SHA256SUMS.sig && \
    gpg --batch --verify packer_${VERSION}_SHA256SUMS.sig packer_${VERSION}_SHA256SUMS && \
    grep packer_${VERSION}_linux_${packerArch}.zip packer_${VERSION}_SHA256SUMS | sha256sum -c && \
    unzip -d /tmp/build packer_${VERSION}_linux_${packerArch}.zip && \
    cp /tmp/build/packer /bin/packer && \
    cd /tmp && \
    rm -rf /tmp/build && \
    gpgconf --kill all && \
    apk del gnupg openssl && \
    rm -rf /root/.gnupg && \
    # Tiny smoke test to ensure the binary we downloaded runs
    packer version

ENTRYPOINT ["/bin/packer"]


# Light docker image which can be used to run the binary from a container.
# This image builds from the locally generated binary in ./bin/, and from CI-built binaries within CI. 
# To generate the local binary, run `make dev`.
# This image is published to DockerHub under the `light`, `light-$VERSION`, and `latest` tags.
FROM docker.mirror.hashicorp.services/alpine:latest as release-light

ARG VERSION
ARG BIN_NAME

# TARGETARCH and TARGETOS are set automatically when --platform is provided.
ARG TARGETOS TARGETARCH

LABEL name="Packer" \
      maintainer="HashiCorp Packer Team <packer@hashicorp.com>" \
      vendor="HashiCorp" \
      version=$VERSION \
      release=$VERSION \
      summary="Packer is a tool for creating identical machine images for multiple platforms from a single source configuration." \
      description="Packer is a tool for creating identical machine images for multiple platforms from a single source configuration. Please submit issues to https://github.com/hashicorp/packer/issues"

RUN apk add --no-cache git bash wget openssl gnupg

COPY dist/$TARGETOS/$TARGETARCH/$BIN_NAME /bin/

ENTRYPOINT ["/bin/packer"]


# Set default target to 'dev'.
FROM dev