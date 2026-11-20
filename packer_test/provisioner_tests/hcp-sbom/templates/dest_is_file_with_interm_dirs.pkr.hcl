# Copyright IBM Corp. 2024, 2026
# SPDX-License-Identifier: BUSL-1.1

packer {
  required_plugins {
    docker = {
      version = ">= 1.0.0"
      source  = "github.com/hashicorp/docker"
    }
  }
}

source "docker" "ubuntu" {
  image  = "ubuntu:20.04"
  commit = true
}

build {
  sources = ["source.docker.ubuntu"]

  provisioner "file" {
    source      = "${path.root}/../testdata/sbom_cyclonedx.json"
    destination = "/tmp/sbom_cyclonedx.json"
  }

  provisioner "hcp-sbom" {
    source      = "/tmp/sbom_cyclonedx.json"
    destination = "./sbom/sbom_cyclonedx"
  }
}
