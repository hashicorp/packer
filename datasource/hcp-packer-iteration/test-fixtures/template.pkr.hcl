source "null" "example" {
  communicator = "none"
}

data "hcp-packer-iteration" "hardened-source" {
  bucket_name = "hardened-ubuntu-16-04"
  channel = "packer-acc-test"
}

data "hcp-packer-image" "aws" {
  bucket_name = "hardened-ubuntu-16-04"
  iteration_id = "${data.hcp-packer-iteration.hardened-source.id}"
  cloud_provider = "aws"
  region = "us-east-1"
}

locals {
  foo              = "${data.hcp-packer-iteration.hardened-source.id}"
  bar              = "${data.hcp-packer-image.aws.id}"
}

build {
  name = "mybuild"
  sources = [
    "source.null.example"
  ]
  provisioner "shell-local" {
    inline = [
      "echo data is ${local.foo}",
      "echo data is ${local.bar}"
    ]
  }
}
