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
  description = "HCP Packer bucket name (must exist or be creatable in your org/project)."
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
  name    = "sbom-test"
  sources = ["source.docker.ubuntu"]

  # This publishes the artifact metadata to HCP Packer Registry.
  hcp_packer_registry {
    bucket_name = var.hcp_bucket_name
  }

  # Automatically generate SBOM using the embedded Syft SDK in the Packer
  # binary and upload to HCP Packer.
  #
  # Packer automatically selects the right binary for the remote host:
  #   - Same OS/arch as host  → uses the running Packer binary (zero cost)
  #   - Different OS/arch, release build → downloads from releases.hashicorp.com
  #   - Different OS/arch, dev build     → cross-compiles from source via `go build`
  #
  # You can also pin a specific binary with `scanner_binary_path` (e.g. for
  # air-gapped environments or pre-built cross-compiled binaries).
  provisioner "hcp-sbom" {
    auto_generate = true
    scan_path     = "/"
    destination   = "./sbom" # local folder must already exist
    sbom_name     = "auto-sbom"
    scanner_args = ["-o", "spdx-json", "-q"]
    # {{.Path}} is the uploaded Packer binary; sbom-generate is the subcommand.
    execute_command = "chmod +x {{.Path}} && {{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}"
  }
}