packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "freebsd" {
  ami_name      = "sbom-amzn-freebsd-no-exec"
  instance_type = "t3.large"
  region        = "us-west-2"

  source_ami_filter {
    filters = {
      name                = "FreeBSD 13.4-RELEASE-amd64*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["782086452779"]
  }

  ssh_username = "ec2-user"
}

hcp_packer_registry {
  bucket_name = "native-sbom"
  description = "Amazon FreeBSD SBOM test without execute_command override."
}

build {
  name    = "sbom-amazon-freebsd-no-exec"
  sources = ["source.amazon-ebs.freebsd"]

  provisioner "shell" {
    inline = ["echo 'packer works on amazon freebsd'"]
  }

  provisioner "hcp-sbom" {
    auto_generate = true
    scan_path     = "/usr/bin"
    destination   = "sbom.json"
    sbom_name     = "amzn-freebsd-no-exec"
    scanner_args  = ["-o", "cyclonedx-json"]
  }
}
