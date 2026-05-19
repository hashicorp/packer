packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
}
source "amazon-ebs" "ubuntu" {
  ami_name      = "test17"
  instance_type = "t2.large"
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
  # skip_create_ami = true
}
hcp_packer_registry {
  bucket_name = "native-sbom"
  description = <<EOT
This registry contains Packer plugins that generate SBOMs for native images. The plugins are designed to work with the HCP Packer plugin system and can be used to create SBOMs for various types of native images, including those created with Amazon EBS.
  EOT
}
build {
  name = "native-sbom-packer"
  sources = [
    "source.amazon-ebs.ubuntu"
  ]
  provisioner "shell" {
    inline = ["echo 'packer works'"]
  }
  provisioner "hcp-sbom" {
    auto_generate = true
    scan_path = "/usr/bin"
    destination = "sbom.json"
  }
  # provisioner "breakpoint" {
  # }
}
