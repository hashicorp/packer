# Copyright IBM Corp. 2026
# SPDX-License-Identifier: BUSL-1.1

packer {
  required_plugins {
    docker = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/docker"
    }
  }
}

source "docker" "cnspec" {
  image   = "mondoo/cnspec:latest"
  discard = true
}

build {
  sources = ["source.docker.cnspec"]

  provisioner "shell" {
    inline = ["cnspec sbom local --output cyclonedx-json --output-target /tmp/sbom_cyclonedx.json"]
  }

  provisioner "hcp-sbom" {
    source      = "/tmp/sbom_cyclonedx.json"
    destination = "./sbom/sbom_cyclonedx.json"
  }
}
