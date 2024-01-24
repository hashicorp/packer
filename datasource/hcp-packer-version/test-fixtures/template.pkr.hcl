source "null" "example" {
  communicator = "none"
}

data "hcp-packer-version" "hardened-source" {
  bucket_name = "simple"
  channel_name = "latest"
}

data "hcp-packer-artifact" "file" {
  bucket_name = "simple"
  version_fingerprint = "${data.hcp-packer-version.hardened-source.fingerprint}"
  platform = "packer.file"
  region = %q
}

locals {
  foo              = "${data.hcp-packer-version.hardened-source.id}"
  bar              = "${data.hcp-packer-artifact.file.external_identifier}"
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
