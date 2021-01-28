source "amazon-ebs" "example" {
  communicator = "none"
  source_ami = "potato"
  ami_name = "potato"
  instance_type = "potato"
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

  provisioner "comment-my-provisioner" {

  }
  provisioner "shell-local" {
    inline = ["yo"]
  }
}
