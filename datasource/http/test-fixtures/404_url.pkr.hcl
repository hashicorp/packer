
source "null" "example" {
  communicator = "none"
}

data "http" "basic" {
  url = "https://developer.hashicorp.com/packer/thisWillFail"
}

locals {
  url = "${data.http.basic.url}"
}

build {
  name = "mybuild"
  sources = [
    "source.null.example"
  ]
  provisioner "shell-local" {
    inline = [
      "echo data is ${local.url}",
    ]
  }
}
