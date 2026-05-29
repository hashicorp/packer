packer {
  required_plugins {
    docker = {
      source  = "github.com/hashicorp/docker"
      version = ">= 1.0.0"
    }
  }
}

variable "hcp_bucket_name" {
  type        = string
  default     = "sbom-bucket-test"
  description = "HCP Packer bucket name."
}

variable "image_name" {
  type        = string
  default = "localhost:5000/huge-sbom-image"
  description = "Docker image to build from for large SBOM generation."
}

source "docker" "ubuntu" {
  image  = var.image_name
  commit = true
}

build {
  name    = "sbom-test"
  sources = ["source.docker.ubuntu"]

  hcp_packer_registry {
    bucket_name = var.hcp_bucket_name
  }

  provisioner "hcp-sbom" {
    auto_generate = true
    scan_path     = "/"
    destination   = "./sbom"
    sbom_name     = "auto-sbom"
    scanner_args  = ["-o", "spdx-json"]
    execute_command = "chmod +x {{.Path}} && {{.Path}} sbom-generate {{.Args}} {{.ScanPath}} > {{.Output}}"
  }
}