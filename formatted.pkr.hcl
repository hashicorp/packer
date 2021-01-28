source "amazon-ebs" "example" {
  communicator = "none"
  ami_name = "potato"
}

build {
  sources = ["source.amazon-ebs.example"]
}

packer {
  required_plugins {
    comment = {
      source  = "sylviamoss/comment"
      version = "v0.2.15"
    }
  }
}

build {
  sources = ["source.amazon-ebs.example"]

  provisioner "shell-local" {
    inline = ["yo"]
  }
}
