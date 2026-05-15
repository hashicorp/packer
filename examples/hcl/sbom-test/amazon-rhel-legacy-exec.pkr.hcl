packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "rhel" {
  ami_name      = "sbom-amzn-rhel-legacy"
  instance_type = "t3.large"
  region        = "us-west-2"

  source_ami_filter {
    filters = {
      name                = "RHEL-9.*_HVM-*-x86_64-*-Hourly2-GP3"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["309956199498"]
  }

  ssh_username = "ec2-user"
}

hcp_packer_registry {
  bucket_name = "native-sbom"
  description = "Amazon RHEL SBOM test with legacy execute_command style."
}

build {
  name    = "sbom-amazon-rhel-legacy-exec"
  sources = ["source.amazon-ebs.rhel"]

  provisioner "shell" {
    inline = ["echo 'packer works on amazon rhel legacy'"]
  }

  provisioner "hcp-sbom" {
    auto_generate = true
    scan_path     = "/usr/bin"
    destination   = "sbom.json"
    sbom_name     = "amzn-rhel-legacy"
    scanner_args  = ["-o", "cyclonedx-json"]

    # Legacy style, compatibility path should auto-inject sbom-generate.
    execute_command = "chmod +x {{.Path}} && {{.Path}} {{.Args}} {{.ScanPath}} > {{.Output}}"
  }
}
