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
  default     = "ubuntu:22.04"
  description = "Base docker image to build from."
}

source "docker" "ubuntu" {
  image  = var.image_name
  commit = true
}

build {
  name    = "sbom-docker-linux-legacy-exec"
  sources = ["source.docker.ubuntu"]

  hcp_packer_registry {
    bucket_name = var.hcp_bucket_name
  }

  provisioner "hcp-sbom" {
    auto_generate = true
    scan_path     = "/"
    destination   = "./sbom"
    sbom_name     = "docker-linux-legacy"
    scanner_args  = ["-o", "cyclonedx-json"]

    # Legacy style command that omits 'sbom-generate'.
    # This should now work through compatibility injection in code.
    execute_command = "chmod +x {{.Path}} && {{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}"
  }
}
