source "null" "example" {
  communicator = "none"
}

data "hcp-packer-iteration" "hardened-source" {
  bucket_name = "simple-deprecated"
  channel = "latest"
}

data "hcp-packer-image" "file" {
  bucket_name = "simple-deprecated"
  iteration_id = "${data.hcp-packer-iteration.hardened-source.id}"
  cloud_provider = "packer.file"
  region = %q
}

locals {
  foo              = "${data.hcp-packer-iteration.hardened-source.id}"
  bar              = "${data.hcp-packer-image.file.id}"
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
