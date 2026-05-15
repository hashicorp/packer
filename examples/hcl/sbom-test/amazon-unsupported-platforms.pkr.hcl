packer {
  required_plugins {
    amazon = {
      version = ">= 1.2.8"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

# This template captures scenarios expected to fail SBOM generation on targets
# where the embedded Syft generator is not supported by build tags.
# Current unsupported targets in code include: netbsd, openbsd, solaris,
# mips, mipsle, mips64, and freebsd/386.

source "amazon-ebs" "freebsd_i386_candidate" {
  ami_name      = "sbom-amzn-freebsd-unsupported"
  instance_type = "t3.large"
  region        = "us-west-2"

  # Candidate filter for historical FreeBSD i386 AMIs (may not resolve in every region/account).
  source_ami_filter {
    filters = {
      name                = "FreeBSD*-i386*"
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
  description = "Amazon SBOM test for expected unsupported platform targets."
}

build {
  name    = "sbom-amazon-unsupported-platforms"
  sources = ["source.amazon-ebs.freebsd_i386_candidate"]

  provisioner "shell" {
    inline = ["echo 'attempting unsupported SBOM target'"]
  }

  provisioner "hcp-sbom" {
    auto_generate = true
    scan_path     = "/usr/bin"
    destination   = "sbom.json"
    sbom_name     = "amzn-unsupported-platform"
    scanner_args  = ["-o", "cyclonedx-json"]
  }
}
