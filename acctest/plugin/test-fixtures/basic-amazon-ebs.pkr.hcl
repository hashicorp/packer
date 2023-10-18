packer {
  required_plugins {
    amazon = {
      version = "~> 1"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "basic-test" {
  region          = "us-east-1"
  instance_type   = "m3.medium"
  source_ami      = "ami-76b2a71e"
  ssh_username    = "ubuntu"
  ami_name        = "packer-plugin-amazon-ebs-test"
  skip_create_ami = true
}

build {
  sources = ["source.amazon-ebs.basic-test"]
}
