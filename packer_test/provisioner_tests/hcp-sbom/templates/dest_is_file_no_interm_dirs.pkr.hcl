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

  provisioner "shell" {
    inline = [
      "apt-get update -y",
      "apt-get install -y curl",
      "bash -c \"$(curl -sSL https://install.mondoo.com/sh)\""
    ]
  }

  provisioner "shell" {
    inline = [
      "cnquery sbom --output cyclonedx-json --output-target /tmp/sbom_cyclonedx.json",
    ]
  }

  provisioner "hcp-sbom" {
    source      = "/tmp/sbom_cyclonedx.json"
    destination = "./sbom/sbom_cyclonedx.json"
  }
}
