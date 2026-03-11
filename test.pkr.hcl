packer {
  required_plugins {
    docker = {
      version = ">= 1.1.0"
      source  = "github.com/hashicorp/docker"
    }
  }
}

# HCP Packer registry — provisioner blocks below will be
# automatically published as enforced blocks to this bucket.
hcp_packer_registry {
  bucket_name = "ubuntu-test"
  description = "Test Ubuntu image with enforced provisioners"

  bucket_labels = {
    "team"    = "platform"
    "os"      = "ubuntu"
    "purpose" = "testing"
  }
}

source "docker" "ubuntu" {
  image  = "ubuntu:22.04"
  commit = true
}

build {
  name = "ubuntu-test"

  sources = ["source.docker.ubuntu"]

  provisioner "shell" {
    inline = [
      "apt-get update -y",
      "apt-get install -y curl wget jq"
    ]
  }

  provisioner "shell" {
    inline = [
      "echo 'Creating app user...'",
      "useradd -m -s /bin/bash appuser",
      "mkdir -p /opt/app",
      "chown appuser:appuser /opt/app"
    ]
  }

  provisioner "shell" {
    inline = [
      "echo 'Applying security hardening...'",
      "echo 'net.ipv4.ip_forward = 0' >> /etc/sysctl.conf",
      "echo 'Build complete!'"
    ]
  }
}
