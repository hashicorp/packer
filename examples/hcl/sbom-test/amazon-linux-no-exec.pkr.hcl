packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "ubuntu" {
  ami_name      = "sbom-amzn-linux-no-exec"
  instance_type = "t3.large"
  region        = "us-west-2"

  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-jammy-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }

  ssh_username = "ubuntu"
}

hcp_packer_registry {
  bucket_name = "native-sbom"
  description = "Amazon Linux SBOM test without execute_command override."
}

build {
  name    = "sbom-amazon-linux-no-exec"
  sources = ["source.amazon-ebs.ubuntu"]

  provisioner "shell" {
    inline = ["echo 'packer works on amazon linux'"]
  }

  provisioner "hcp-sbom" {
    auto_generate = true
    scan_path     = "/usr/bin"
    destination   = "sbom.json"
    sbom_name     = "amzn-linux-no-exec"
    scanner_args  = ["-o", "cyclonedx-json"]
  }
}
