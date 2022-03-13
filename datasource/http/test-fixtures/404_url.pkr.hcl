
source "null" "example" {
  communicator = "none"
}

data "http" "basic" {
  url = "https://www.packer.io/thisWillFail"
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
